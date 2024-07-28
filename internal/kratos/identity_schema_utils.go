package kratos

import (
	"context"
	"github.com/lus/hydra-consent/internal/static"
	oryKratos "github.com/ory/client-go"
	oryHydra "github.com/ory/hydra-client-go"
	"strings"
)

func ExtractSessionValues(ctx context.Context, client *oryKratos.APIClient, scopes []string, identity *oryKratos.Identity) (*oryHydra.AcceptOAuth2ConsentRequestSession, error) {
	// Retrieve the raw schema data directly from Kratos
	schema, _, err := client.IdentityAPI.GetIdentitySchema(ctx, identity.SchemaId).Execute()
	if err != nil {
		return nil, err
	}

	// Retrieve the actual trait values of the identity
	traitMap, ok := identity.Traits.(map[string]any)
	if !ok {
		return nil, nil
	}

	accessTokenValues := make(map[string]any)
	idTokenValues := make(map[string]any)

	// Process every trait of the schema
	traits, ok := extractNestedValue[map[string]any](schema, "properties.traits.properties")
	if !ok {
		return nil, nil
	}
	for trait, rawValues := range traits {
		// Check if the trait schema has the correct format
		values, ok := rawValues.(map[string]any)
		if !ok {
			continue
		}

		// Try to retrieve the actual value of the corresponding trait from the identity
		traitValue, ok := traitMap[trait]
		if !ok {
			continue
		}

		// If the schema defines that a scope is required for this value, check if it was granted to the client
		requiredScope, _ := extractNestedValue[string](values, static.KratosIdentitySchemaExtensionKey+".required_scope")
		if requiredScope != "" && !contains(scopes, requiredScope) {
			continue
		}

		// If the schema defines a key for the trait for the ID token, add it to the session
		idTokenPath, ok := extractNestedValue[string](values, static.KratosIdentitySchemaExtensionKey+".session_data.id_token_key")
		if ok {
			idTokenValues[idTokenPath] = traitValue
		}

		// If the schema defines a key for the trait for the access token, add it to the session
		accessTokenPath, ok := extractNestedValue[string](values, static.KratosIdentitySchemaExtensionKey+".session_data.access_token_key")
		if ok {
			accessTokenValues[accessTokenPath] = traitValue
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

func contains(arr []string, val string) bool {
	for _, elem := range arr {
		if elem == val {
			return true
		}
	}
	return false
}
