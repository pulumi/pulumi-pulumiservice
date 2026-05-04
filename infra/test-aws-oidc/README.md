# test-aws-oidc

One-shot Pulumi program that provisions the AWS-side trust the Insights
example tests rely on:

- An IAM OIDC identity provider for Pulumi Cloud's issuer.
- An IAM role whose trust policy allows
  `sts:AssumeRoleWithWebIdentity` from that provider, gated on the ESC-issued
  `aud` and `sub` claims.
- A managed policy attachment so the role can actually read AWS resources
  during a scan (defaults to `ReadOnlyAccess`).

This is **not** a test fixture. It is documentation-as-code for the trust
that the following examples assume:

- `examples/yaml-insights-account/`
- `examples/yaml-policy-groups-accounts/`
- `examples/ts-insights-account-invokes/`

Those examples currently hard-code the AWS docs placeholder account
(`123456789012`) and a role named `PulumiInsightsRole`. After
[pulumi-service#41970][validate-pr] the placeholder is rejected at create
time, so the examples need to point at a real role in a real account.

[validate-pr]: https://github.com/pulumi/pulumi-service/pull/41970

## Prerequisites

- AWS credentials with permission to create an IAM OIDC provider and IAM
  role (e.g. `iam:CreateOpenIDConnectProvider`, `iam:CreateRole`,
  `iam:AttachRolePolicy`).
- A Pulumi Cloud organization that the Insights tests run against. The
  examples currently default to `service-provider-test-org`.

## Configure

Defaults are set to match the values observed in the failing CI run
(staging backend, `service-provider-test-org`, `aws:service-provider-test-org`
audience). For that environment, no config is required.

```sh
cd infra/test-aws-oidc
pulumi stack init dev
```

Override any of the defaults if you're targeting a different org or
backend:

| Config key      | Default                                   | Notes                                                                                                                     |
| --------------- | ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| `orgName`       | `service-provider-test-org`               | Pulumi Cloud org that issues the JWT. Embedded in the `sub` claim.                                                        |
| `issuerHost`    | `api.pulumi-staging.io/oidc`              | Issuer host+path, no scheme. Used as the IAM condition key prefix and as the path of the OIDC provider URL. Set to `api.pulumi.com/oidc` for production. |
| `audience`      | `aws:service-provider-test-org`           | The `aud` claim ESC mints. Verified against the failing run's `Audience: "aws:service-provider-test-org"` log line.       |
| `roleName`      | `PulumiInsightsRole`                      | Match whatever the ESC env's `roleArn` references.                                                                        |
| `scanPolicyArn` | `arn:aws:iam::aws:policy/ReadOnlyAccess`  | Tighten if you don't want the scan role to have account-wide read.                                                        |

## Apply

```sh
pulumi up
pulumi stack output roleArn
```

## Wire up the examples

Take the `roleArn` output and drop it into the three example ESC
environments in place of the placeholder
`arn:aws:iam::123456789012:role/PulumiInsightsRole`:

- `examples/yaml-insights-account/Pulumi.yaml`
- `examples/yaml-policy-groups-accounts/Pulumi.yaml`
- `examples/ts-insights-account-invokes/index.ts`

For CI, prefer threading the role ARN through stack config (`aws:roleArn`
or similar) rather than hard-coding the new value, so it stays out of the
checked-in examples and can vary by environment.

## Caveats

- `thumbprintLists` uses the well-known dummy
  (`ffffffffffffffffffffffffffffffffffffffff`). AWS no longer enforces
  thumbprint validation for IAM OIDC providers backed by certs in its
  trusted root list. Pin a real thumbprint if your security policy
  requires it.
- The trust policy's `sub` claim wildcards on env name
  (`pulumi:environments:org:<orgName>:env:*`). Tighten to a specific
  project/env if you want per-test scoping.
- The `aud` and `sub` claim formats here match Pulumi Cloud's documented
  ESC `aws-login.oidc` behavior at the time of writing. If the JWT
  Pulumi mints in your environment has different claims, run a failed
  assume once and inspect the token (or check Pulumi Cloud OIDC docs)
  to recover the exact values, then update the conditions.
