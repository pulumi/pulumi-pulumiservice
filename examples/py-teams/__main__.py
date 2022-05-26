"""A Python Pulumi Service Teams program"""

from pulumi_pulumiservice import Team
from pulumi_random import RandomString

rand = RandomString("team-suffix", length=5, special=False)

team = Team(
    "team",
    name=rand.result.apply(lambda res: f"brand-new-python-team-{res}"),
    organization_name="service-provider-test-org",
    display_name="PulumiUP Team",
    team_type="pulumi",
    members=["pulumi-bot", "service-provider-example-user"],
)
