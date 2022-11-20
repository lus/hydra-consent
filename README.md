# hydra-consent &nbsp; [![](https://github.com/lus/hydra-consent/actions/workflows/docker.yml/badge.svg)](https://github.com/lus/hydra-consent/actions/workflows/docker.yml)

`hydra-consent` is a simple yet powerful consent flow handler for [Ory Hydra](https://github.com/ory/hydra).

Additionally, it natively integrates with [Ory Kratos](https://github.com/ory/kratos) to dynamically propagate identity
traits to ID & session tokens issued by Hydra. 

**Note:** This service works best in environments using both Hydra and Kratos and only creating Hydra clients for
internal services, like an SSO-only environment, as it does not yet provide a user-facing UI for manually approving
requested scopes. Consent challenges will only be auto-accepted or auto-denied.
It also only supports trait propagation for Kratos.

## Installation & Configuration 

### Docker

A Docker image is automatically built and pushed to GHCR.
A `docker-compose.yml` file may look like this:

```yml
services:
    hydra-consent:
        image: ghcr.io/lus/hydra-consent:latest
        restart: unless-stopped
        ports:
            - 8080:8080
        environment:
            # By default the application runs in development mode.
            # I highly recommend setting this value for production usage.
            ENVIRONMENT: prod
            # The minimum log level to print.
            # Available values: 'trace', 'debug', 'info', 'warn', 'error', 'fatal', 'panic' and 'disabled'
            # The default value is 'info'. I recommend leaving this as this will not clutter your console anyway.
            LOG_LEVEL: info
            # The address to bind the HTTP server to.
            # The default value is :8080 (= 0.0.0.0:8080).
            LISTEN_ADDRESS: :8080
            # The address to Hydra's admin API.
            # DO NOT expose this to the public without proper authentication & authorization.
            # This field is required.
            HYDRA_ADMIN_API: http://hydra:4445
            # The address to Kratos' admin API.
            # DO NOT expose this to the public without proper authentication & authorization.
            # This field is optional. Only specify it if you want to enable the native Kratos trait propagation.
            KRATOS_ADMIN_API: http://kratos:4434
```

Please note that for this exact configuration to work, the container has to be in at least one mutual Docker network
with Hydra and Kratos in order to be able to reach their hopefully not publicly exposed admin APIs. 

### From source

After making sure that [Go 1.19](https://go.dev/dl/) is installed, simply follow these steps:

1. Clone the repository and enter the directory it got cloned to:
    ```shell
    git clone https://github.com/lus/hydra-consent && cd hydra-consent
    ```
2. Build the binary:
    ```shell
    go build -o server cmd/server/main.go
    ```
3. Make sure to set the environment variables according to your configuration. A `.env` file is supported natively:
    ```
    # By default the application runs in development mode.
    # I highly recommend setting this value for production usage.
    ENVIRONMENT: prod
    
    # The minimum log level to print.
    # Available values: 'trace', 'debug', 'info', 'warn', 'error', 'fatal', 'panic' and 'disabled'
    # The default value is 'info'. I recommend leaving this as this will not clutter your console anyway.
    LOG_LEVEL: info
    
    # The address to bind the HTTP server to.
    # The default value is :8080 (= 0.0.0.0:8080).
    LISTEN_ADDRESS: :8080
    
    # The address to Hydra's admin API.
    # DO NOT expose this to the public without proper authentication & authorization.
    # This field is required.
    HYDRA_ADMIN_API: http://hydra:4445
    
    # The address to Kratos' admin API.
    # DO NOT expose this to the public without proper authentication & authorization.
    # This field is optional. Only specify it if you want to enable the native Kratos trait propagation.
    KRATOS_ADMIN_API: http://kratos:4434
    ```
4. Run the application:
    ```shell
    ./server
    ```

## Getting started

### Set Hydra's consent URL

First of all, you have to configure Hydra to redirect users to this service for consent challenges.
Do that by setting the `urls.consent` value in Hydra's configuration file to the `/consent` endpoint exposed by
this service.

### Trust clients

After correctly configuring your setup, this service will deny every consent challenge by default.
This is because you did not mark any OAuth2 clients as trusted yet.

Set the `trusted` metadata of every client whose consent challenges should be accepted immediately to `true`.

**`hydra-consent` does not implement any user-facing UI for manually approving requested scopes just yet!**

### Done!

Congratulations! You successfully set up `hydra-consent`! Yes, it was that easy.

However, if you use Ory Kratos as your identity provider, you may want to include some identity traits in the ID and/or
session token(s). The following section will explain how this is possible.

## Kratos identity trait propagation

**Note:** In order for this to work, do not forget to set the `KRATOS_ADMIN_API` environment variable.

`hydra-consent` natively integrates with Ory Kratos in order to be able to inject identity traits into the OAuth2 & OIDC
ID & session tokens.
It will search for values set in the `lus/hydra-consent` extension field for every trait defined in the
**identity schema**.

Example schema making use of every feature supported by this service right now:

```jsonc
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "$id": "https://example.com/schemas/user.schema.json",
    "title": "User",
    "type": "object",
    "properties": {
        "traits": {
            "type": "object",
            "properties": {
                "username": {
                    "title": "Username",
                    "type": "string",
                    "minLength": 1,
                    "maxLength": 20,
                    "pattern": "^[a-z0-9]+$",
                    "ory.sh/kratos": {
                        "credentials": {
                            "password": {
                                "identifier": true
                            }
                        }
                    },
                    "lus/hydra-consent": {
                        "required_scope": "profile",
                        "session_data": {
                            "id_token_key": "preferred_username"
                            "session_token_key": "username"
                        }
                    }
                },
                "email": {
                    "title": "E-Mail",
                    "type": "string",
                    "format": "email",
                    "ory.sh/kratos": {
                        "credentials": {
                            "password": {
                                "identifier": true
                            }
                        },
                        "recovery": {
                            "via": "email"
                        },
                        "verification": {
                            "via": "email"
                        }
                    },
                    "lus/hydra-consent": {
                        "required_scope": "email",
                        "session_data": {
                            "id_token_key": "email"
                        }
                    }
                }
            },
            "required": [
                "username",
                "email"
            ],
            "additionalProperties": false
        }
    }
}
```

Pay attention to the `lus/hydra-consent` schema extension fields.

If the `required_scope` field is present and a string, the corresponding trait will only be propagated to clients
granted this scope. If it is missing, it will be propagated to every client.

The `session_data.id_token_key` and `session_data.session_token_key` fields specify the keys in the ID and session token
under which the trait value will be set.
These fields are also optional, not setting one of them simply will not include the trait in the corresponding token.

## Support

Feel free to open issues in this repository if you encounter any problem or want to suggest a feature.
If you want to ask a quick question, feel free to join my [Discord server](https://go.lus.pm/discord).
