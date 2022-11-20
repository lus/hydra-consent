package kratos

import (
	"context"
	"github.com/lus/hydra-consent/internal/static"
	oryKratos "github.com/ory/client-go"
	oryHydra "github.com/ory/hydra-client-go"
	"strings"
)

func ExtractSessionValues(ctx context.Context, client *oryKratos.APIClient, identity *oryKratos.Identity) (*oryHydra.AcceptOAuth2ConsentRequestSession, error) {
	schema, _, err := client.V0alpha2Api.GetIdentitySchema(ctx, identity.SchemaId).Execute()
	if err != nil {
		return nil, err
	}

	traitMap, ok := identity.Traits.(map[string]any)
	if !ok {
		return nil, nil
	}

	accessTokenValues := make(map[string]any)
	idTokenValues := make(map[string]any)

	traits, ok := extractNestedValue[map[string]any](schema, "properties.traits.properties")
	if !ok {
		return nil, nil
	}
	for trait, rawValues := range traits {
		values, ok := rawValues.(map[string]any)
		if !ok {
			continue
		}
		idTokenPath, ok := extractNestedValue[string](values, static.KratosIdentitySchemaExtensionKey+".session_data.id_token_path")
		if ok {
			traitValue, ok := traitMap[trait]
			if ok {
				idTokenValues[idTokenPath] = traitValue
			}
		}
		accessTokenPath, ok := extractNestedValue[string](values, static.KratosIdentitySchemaExtensionKey+".session_data.access_token_path")
		if ok {
			traitValue, ok := traitMap[trait]
			if ok {
				accessTokenValues[accessTokenPath] = traitValue
			}
		}
	}

	return &oryHydra.AcceptOAuth2ConsentRequestSession{
		AccessToken: accessTokenValues,
		IdToken:     idTokenValues,
	}, nil
}

func extractNestedValue[T any](structure map[string]any, key string) (T, bool) {
	var defaultValue T
	keys := strings.Split(key, ".")
	currentMap := structure
	for i := 0; i < len(keys)-1; i++ {
		newMap, ok := currentMap[keys[i]].(map[string]any)
		if !ok {
			return defaultValue, ok
		}
		currentMap = newMap
	}
	val, ok := currentMap[keys[len(keys)-1]].(T)
	return val, ok
}
