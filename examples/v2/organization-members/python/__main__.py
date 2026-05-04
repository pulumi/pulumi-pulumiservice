import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
member_login = config.get("memberLogin") or "pulumi-bot"
member_role = config.get("memberRole") or "member"

member = ps_v2.OrganizationMember(
    "member",
    org_name=service_org,
    user_login=member_login,
    role=member_role,
)

pulumi.export("memberId", member.id)
