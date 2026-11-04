import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const serviceOrg = "service-provider-test-org2";

// Thumbprints must match the certificate the issuer currently serves, so they
// are supplied via config rather than hardcoded. Compute one with:
//   openssl s_client -connect <issuer-host>:443 </dev/null | openssl x509 -fingerprint -sha256 -noout
const config = new pulumi.Config();
// Tests override the issuer URLs so that concurrent runs don't conflict;
// see oidcIssuerTestConfig in examples/examples_utils_test.go.
const pulumiIssuerUrl = config.get("pulumiIssuerUrl") ?? "https://api.pulumi.com/oidc";
const githubIssuerUrl = config.get("githubIssuerUrl") ?? "https://token.actions.githubusercontent.com";
const pulumiThumbprint = config.require("pulumiThumbprint");
const githubThumbprint = config.require("githubThumbprint");

// A Pulumi OIDC Issuer with a basic policy
const pulumiOidcIssuer = new service.OidcIssuer("pulumi_issuer", {
  organization: serviceOrg,
  name: "pulumi_issuer",
  url: pulumiIssuerUrl,
  thumbprints: [pulumiThumbprint],
  policies: [
    {
      decision: "allow",
      rules: {
          "aud": "urn:pulumi:org:"+serviceOrg,
          "sub": "pulumi:deploy:org:myTestOrg:project:myTestProject:*"
      },
      userLogin: "pulumipus",
      tokenType: "personal"
    }
  ]
})

// A Github OIDC Issuer with all types of policies and TLS certificate thumbprint
const githubOidcIssuer = new service.OidcIssuer("github_issuer", {
  organization: serviceOrg,
  name: "github_issuer",
  url: githubIssuerUrl,
  thumbprints: [githubThumbprint],
  policies: [
    {
      decision: "deny",
      rules: {
          "aud": "urn:pulumi:org:"+serviceOrg,
          "sub": "repo:organization/repo:*"
      },
      userLogin: "pulumipus",
      tokenType: "personal"
    },
    {
      decision: "allow",
      rules: {
          "aud": "urn:pulumi:org:"+serviceOrg,
          "sub": "repo:organization/repo:*"
      },
      authorizedPermissions: [
        "admin"
      ],
      tokenType: "organization"
    },
    {
      decision: "deny",
      rules: {
          "aud": "urn:pulumi:org:"+serviceOrg,
          "sub": "repo:organization/repo:*"
      },
      teamName: "dream-team",
      tokenType: "team"
    }
  ]
})

