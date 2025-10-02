import * as service from "@pulumi/pulumiservice";

const serviceOrg = "service-provider-test-org2";

// A Pulumi OIDC Issuer with a basic policy
const pulumiOidcIssuer = new service.OidcIssuer("pulumi_issuer", {
  organization: serviceOrg,
  name: "pulumi_issuer",
  url: "https://api.pulumi.com/oidc",
  thumbprints: [
    "57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da"
  ],
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
  url: "https://token.actions.githubusercontent.com",
  thumbprints: [
    "caef08400c87bedb0db28f1a0dc0b49e658c8e90a985b8c3e6a6e7f51a2d09d7"
  ],
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

