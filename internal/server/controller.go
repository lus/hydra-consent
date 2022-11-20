package server

import (
	"github.com/lus/hydra-consent/internal/kratos"
	"github.com/lus/hydra-consent/internal/ptr"
	"github.com/lus/hydra-consent/internal/static"
	oryKratos "github.com/ory/client-go"
	oryHydra "github.com/ory/hydra-client-go"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
	"time"
)

type controller struct {
	Hydra  *oryHydra.APIClient
	Kratos *oryKratos.APIClient
}

func (cnt *controller) Endpoint(writer http.ResponseWriter, request *http.Request) {
	challengeId := request.URL.Query().Get("consent_challenge")
	if challengeId == "" {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("A consent_challenge is required."))
		return
	}

	challenge, response, err := cnt.Hydra.OAuth2Api.GetOAuth2ConsentRequest(request.Context()).
		ConsentChallenge(challengeId).
		Execute()
	if err != nil {
		if response != nil && response.StatusCode == http.StatusNotFound {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("No consent challenge was found."))
			return
		}
		cnt.error(writer, err)
		return
	}

	// Check if the client requesting consent is marked as trusted
	accept := false
	metadata := challenge.Client.Metadata
	if metadata != nil {
		if metaMap, ok := metadata.(map[string]any); ok {
			trusted, _ := metaMap[static.HydraTrustedClientMetadataKey].(bool)
			accept = trusted
		}
	}

	// Deny the consent challenge if the client is not trusted
	if !accept {
		log.Debug().
			Str("challenge", challengeId).
			Str("reason", "client not trusted").
			Msg("Rejecting a consent challenge...")
		redirect, _, err := cnt.Hydra.OAuth2Api.RejectOAuth2ConsentRequest(request.Context()).
			ConsentChallenge(challengeId).
			RejectOAuth2Request(oryHydra.RejectOAuth2Request{
				Error:            ptr.Ptr("access_denied"),
				ErrorDebug:       ptr.Ptr("lus/hydra-consent"),
				ErrorDescription: ptr.Ptr("The client requesting the consent is not trusted."),
				ErrorHint:        ptr.Ptr("trust client"), // TODO: Link to README section
				StatusCode:       ptr.Ptr(int64(418)),
			}).
			Execute()
		if err != nil {
			cnt.error(writer, err)
			return
		}
		http.Redirect(writer, request, redirect.RedirectTo, http.StatusFound)
		return
	}

	var session *oryHydra.AcceptOAuth2ConsentRequestSession

	if cnt.Kratos != nil {
		// Retrieve the identity of the subject directly from Kratos
		subject := ""
		if challenge.Subject != nil {
			subject = *challenge.Subject
		}
		identity, response, err := cnt.Kratos.V0alpha2Api.AdminGetIdentity(request.Context(), subject).Execute()
		if err != nil {
			if response == nil || response.StatusCode != http.StatusNotFound {
				cnt.error(writer, err)
				return
			}
		}

		// Add the identity's traits to the session (ID & access tokens) according to the schema
		if identity != nil {
			parsedSession, err := kratos.ExtractSessionValues(request.Context(), cnt.Kratos, challenge.RequestedScope, identity)
			if err != nil {
				cnt.error(writer, err)
				return
			}
			session = parsedSession
			if session != nil {
				log.Debug().
					Str("scopes", strings.Join(challenge.RequestedScope, ",")).
					Interface("session", *session).
					Msg("Injecting session data...")
			}
		} else {
			log.Debug().Str("challenge", challengeId).Str("subject", subject).Msg("No Kratos identity was found.")
		}
	}

	// Accept the consent challenge
	log.Debug().Str("challenge", challengeId).Msg("Accepting a consent challenge...")
	redirect, _, err := cnt.Hydra.OAuth2Api.AcceptOAuth2ConsentRequest(request.Context()).
		ConsentChallenge(challengeId).
		AcceptOAuth2ConsentRequest(oryHydra.AcceptOAuth2ConsentRequest{
			GrantAccessTokenAudience: challenge.RequestedAccessTokenAudience,
			GrantScope:               challenge.RequestedScope,
			HandledAt:                ptr.Ptr(time.Now()),
			Session:                  session,
		}).
		Execute()
	if err != nil {
		cnt.error(writer, err)
		return
	}
	http.Redirect(writer, request, redirect.RedirectTo, http.StatusFound)
}

func (cnt *controller) error(writer http.ResponseWriter, err error) {
	writer.WriteHeader(http.StatusInternalServerError)
	writer.Write([]byte("500: " + err.Error()))
	log.Err(err).Msg("The HTTP server encountered an internal error.")
}
