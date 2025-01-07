import * as service from "@pulumi/pulumiservice";

const serviceOrg = "service-provider-test-org2";

// A Pulumi OIDC Issuer with a basic policy
const pulumiOidcIssuer = new service.OidcIssuer("pulumi_issuer", {
  organization: serviceOrg,
  name: "pulumi_issuer",
  url: "https://api.pulumi.com/oidc",
  thumbprints: [
    "df749a0f34ed673f8b0ec898445910c29c170d01d7d34073bd882235974a8a53"
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
    "ac476cca46ea95a04518f08e1e4bd3b010e50587f57706be7e259fd81817f1e1"
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
    },
    {
      decision: "allow",
      rules: {
          "aud": "urn:pulumi:org:"+serviceOrg,
          "sub": "repo:organization/repo:*"
      },
      runnerID: "1234-5678-ABCD-XYZD",
      tokenType: "runner"
    }
  ]
})

