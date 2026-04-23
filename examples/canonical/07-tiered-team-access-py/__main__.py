"""Python variant of canonical/07-tiered-team-access.

Three teams × two stacks × three permission tiers. Behavioral twin of
the sibling YAML program.
"""

import pulumi
from pulumi_pulumiservice.orgs import teams
from pulumi_pulumiservice.stacks import permissions

cfg = pulumi.Config()
organization_name = cfg.get("organizationName") or "service-provider-test-org"
digits = cfg.get("digits") or "00000"

# Team-stack permission levels: 0=none, 101=read, 102=edit, 103=admin.
PERM_READ, PERM_EDIT, PERM_ADMIN = 101, 102, 103

platform_admins = teams.Team(
    "platformAdmins",
    organization_name=organization_name, team_type="pulumi",
    name=f"platform-admins-{digits}",
    display_name="Platform Admins",
    description="Break-glass access to everything. Keep small.",
)
billing_owners = teams.Team(
    "billingOwners",
    organization_name=organization_name, team_type="pulumi",
    name=f"billing-owners-{digits}",
    display_name="Billing Service Owners",
    description="Owns the billing service stacks end-to-end.",
)
developers = teams.Team(
    "developers",
    organization_name=organization_name, team_type="pulumi",
    name=f"developers-{digits}",
    display_name="Developers (all)",
    description="Read everything; deploy nothing without an explicit grant.",
)

# Platform stack grants
permissions.TeamStackPermission(
    "platformAdminPerm",
    organization=organization_name, project="platform", stack="prod",
    team=platform_admins.name, permission=PERM_ADMIN,
)
permissions.TeamStackPermission(
    "platformDevRead",
    organization=organization_name, project="platform", stack="prod",
    team=developers.name, permission=PERM_READ,
)

# Billing stack grants
permissions.TeamStackPermission(
    "billingAdminPerm",
    organization=organization_name, project="billing-service", stack="prod",
    team=platform_admins.name, permission=PERM_ADMIN,
)
permissions.TeamStackPermission(
    "billingOwnerPerm",
    organization=organization_name, project="billing-service", stack="prod",
    team=billing_owners.name, permission=PERM_EDIT,
)
permissions.TeamStackPermission(
    "billingDevRead",
    organization=organization_name, project="billing-service", stack="prod",
    team=developers.name, permission=PERM_READ,
)
