"""Python variant of canonical/01-organization-bootstrap.

Day-0 provisioning: three teams, a team-scoped CI token, a baseline
policy group. Functionally equivalent to the sibling YAML program.
"""

import pulumi
from pulumi_pulumiservice.orgs import teams, tokens, policies

cfg = pulumi.Config()
organization_name = cfg.get("organizationName") or "service-provider-test-org"
digits = cfg.get("digits") or "00000"

admins = teams.Team(
    "admins",
    organization_name=organization_name,
    team_type="pulumi",
    name=f"admins-{digits}",
    display_name="Organization Admins",
    description="Full org control; rotate this membership quarterly.",
)

deployers = teams.Team(
    "deployers",
    organization_name=organization_name,
    team_type="pulumi",
    name=f"deployers-{digits}",
    display_name="CI Deployers",
    description="Automation-only team. Human members discouraged — use the CI token.",
)

teams.Team(
    "readers",
    organization_name=organization_name,
    team_type="pulumi",
    name=f"readers-{digits}",
    display_name="Developers (read-only)",
    description="Default team for new org members; grants stack read access.",
)

ci_token = tokens.TeamAccessToken(
    "ciToken",
    organization_name=organization_name,
    team_name=deployers.name,
    description="Used by GitHub Actions to deploy non-production stacks.",
)

policies.PolicyGroup(
    "defaultGuardrails",
    organization_name=organization_name,
    name=f"baseline-guardrails-{digits}",
)

pulumi.export("ciTokenValue", ci_token.value)
