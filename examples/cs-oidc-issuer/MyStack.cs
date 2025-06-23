using Pulumi;
using Pulumi.PulumiService;
using Pulumi.PulumiService.Inputs;
using System.Collections.Generic;

class MyStack : Pulumi.Stack
{
    public MyStack()
    {
        var serviceOrg = "service-provider-test-org3";

        // A Pulumi OIDC Issuer with a basic policy
        var pulumiOidcIssuer = new OidcIssuer("pulumi_issuer", new OidcIssuerArgs
        {
            Organization = serviceOrg,
            Name = "pulumi_issuer",
            Url = "https://api.pulumi.com/oidc",
            Thumbprints = {
              "df749a0f34ed673f8b0ec898445910c29c170d01d7d34073bd882235974a8a53"
            },
            Policies = {
              new AuthPolicyDefinitionArgs
              {
                Decision = AuthPolicyDecision.Allow,
                Rules = new Dictionary<string, string> {
                    { "aud", "urn:pulumi:org:"+serviceOrg },
                    { "sub", "pulumi:deploy:org:myTestOrg:project:myTestProject:*" },
                },
                UserLogin = "pulumipus",
                TokenType = AuthPolicyTokenType.Personal,
              }
            }
        });

        // A Github OIDC Issuer with all types of policies and TLS certificate thumbprint
        var githubOidcIssuer = new OidcIssuer("github_issuer", new OidcIssuerArgs
        {
          Organization = serviceOrg,
          Name = "github_issuer",
          Url = "https://token.actions.githubusercontent.com",
          Thumbprints = {
            "caef08400c87bedb0db28f1a0dc0b49e658c8e90a985b8c3e6a6e7f51a2d09d7"
          },
          Policies = {
            new AuthPolicyDefinitionArgs
            {
              Decision = AuthPolicyDecision.Deny,
              Rules = new Dictionary<string, string> {
                { "aud", "urn:pulumi:org:"+serviceOrg },
                { "sub", "repo:organization/repo:*" },
              },
              UserLogin = "pulumipus",
              TokenType = AuthPolicyTokenType.Personal
            },
            new AuthPolicyDefinitionArgs
            {
              Decision = AuthPolicyDecision.Allow,
              Rules = new Dictionary<string, string> {
                { "aud", "urn:pulumi:org:"+serviceOrg },
                { "sub", "repo:organization/repo:*" },
              },
              AuthorizedPermissions = {
                AuthPolicyPermissionLevel.Admin
              },
              TokenType = AuthPolicyTokenType.Organization
            },
            new AuthPolicyDefinitionArgs
            {
              Decision = AuthPolicyDecision.Deny,
              Rules = new Dictionary<string, string> {
                { "aud", "urn:pulumi:org:"+serviceOrg },
                { "sub", "repo:organization/repo:*" },
              },
              TeamName = "dream-team",
              TokenType = AuthPolicyTokenType.Team
            },
            new AuthPolicyDefinitionArgs
            {
              Decision = AuthPolicyDecision.Allow,
              Rules = new Dictionary<string, string> {
                { "aud", "urn:pulumi:org:"+serviceOrg },
                { "sub", "repo:organization/repo:*" },
              },
              RunnerID = "1234-5678-ABCD-XYZD",
              TokenType = AuthPolicyTokenType.Runner
            }
          }
        });
    }
}
