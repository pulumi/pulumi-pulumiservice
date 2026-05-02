# OIDC issuer example ported from examples/ts-oidc-issuer to the v2 namespace.
#
# Note: v2 OidcIssuer is auto-derived from the OpenAPI spec. Compared to v1
# (`pulumiservice:OidcIssuer`), input fields differ — v1 exposed `policies`
# and `organization`; v2 exposes `name`, `thumbprints`, `jwks`,
# `maxExpiration`, plus path params (`orgName`). Policy management likely
# lives on a separate resource (or via a richer request body) once the
# polymorphic-union story for RBAC lands.

config serviceOrg "string" {
  default = "service-provider-test-org2"
}

resource pulumiIssuer "pulumiservice:v2:OidcIssuer" {
  orgName     = serviceOrg
  name        = "pulumi_issuer"
  url         = "https://api.pulumi.com/oidc"
  thumbprints = [
    "57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da",
  ]
}

resource githubIssuer "pulumiservice:v2:OidcIssuer" {
  orgName     = serviceOrg
  name        = "github_issuer"
  url         = "https://token.actions.githubusercontent.com"
  thumbprints = [
    "caef08400c87bedb0db28f1a0dc0b49e658c8e90a985b8c3e6a6e7f51a2d09d7",
  ]
  maxExpiration = 3600
}

output pulumiIssuerUrl {
  value = pulumiIssuer.url
}
output githubIssuerUrl {
  value = githubIssuer.url
}
