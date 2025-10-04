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
              "57d3e89f6b25dde3c174dc558e2b2623306a9d81f88a12e8ae7090a86c12f1da"
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
            }
          }
        });
    }
}
