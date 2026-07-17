using Pulumi;
using Pulumi.PulumiService;
using Pulumi.PulumiService.Inputs;
using System.Collections.Generic;

class MyStack : Pulumi.Stack
{
    public MyStack()
    {
        var serviceOrg = "service-provider-test-org3";

        // Thumbprints must match the certificate the issuer currently serves, so they
        // are supplied via config rather than hardcoded. Compute one with:
        //   openssl s_client -connect <issuer-host>:443 </dev/null | openssl x509 -fingerprint -sha256 -noout
        var config = new Pulumi.Config();
        var pulumiThumbprint = config.Require("pulumiThumbprint");
        var githubThumbprint = config.Require("githubThumbprint");

        // A Pulumi OIDC Issuer with a basic policy
        var pulumiOidcIssuer = new OidcIssuer("pulumi_issuer", new OidcIssuerArgs
        {
            Organization = serviceOrg,
            Name = "pulumi_issuer",
            Url = "https://api.pulumi.com/oidc",
            Thumbprints = { pulumiThumbprint },
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
          Thumbprints = { githubThumbprint },
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
