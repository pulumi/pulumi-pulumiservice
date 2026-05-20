import pulumi
import pulumi_pulumiservice.v1 as ps_v1

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
member_login = config.get("memberLogin") or "pulumi-bot"
member_role = config.get("memberRole") or "member"

member = ps_v1.OrganizationMember(
    "member",
    org_name=organization_name,
    user_login=member_login,
    role=member_role,
)

pulumi.export("memberId", member.id)
