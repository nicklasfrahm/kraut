# OIDC Proxy

A simple HTTP server that serves the OIDC discovery document and the JWKs of the Kubernetes API server with a trusted certificate. This allows the service acconts of the Kubernetes cluster to perform an [OAuth 2.0 Token Exchange (RFC8693)][rfc-8693] with external OIDC provider.
