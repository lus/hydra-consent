package static

var (
	// HydraTrustedClientMetadataKey defines the key inside the Hydra client metadata to check for the clients trusted status
	HydraTrustedClientMetadataKey = "trusted"

	// KratosIdentitySchemaExtensionKey defines the key inside the Kratos identity trait schema to take the
	// app-specific values from (id_token_key, access_token_key, required_scope, ...)
	KratosIdentitySchemaExtensionKey = "lus/hydra-consent"
)
