import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
member_login = config.get("memberLogin") or "pulumi-bot"
member_role = config.get("memberRole") or "member"

member = ps_api.OrganizationMember(
    "member",
    org_name=organization_name,
    user_login=member_login,
    role=member_role,
)

pulumi.export("memberId", member.id)
