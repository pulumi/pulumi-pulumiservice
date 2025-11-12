"""A Python Pulumi Service Approval Rules program"""

import pulumi
from pulumi_pulumiservice import Environment, ApprovalRule, EnvironmentIdentifierArgs, ApprovalRuleConfigArgs, EligibleApproverArgs
from pulumi_pulumiservice import TargetActionType, RbacPermission

config = pulumi.Config()

# Create an environment that will be used as the target for our approval rule
environment = Environment(
    "testing-environment",
    organization="service-provider-test-org",
    project="test-project",
    name="testing-environment-approval-py-"+config.require('digits'),
    yaml=pulumi.StringAsset("""values:
  myKey1: myValue1""")
)

# Create an approval rule that governs who can approve updates to the environment
# This rule requires 3 approvals from eligible approvers before any update can proceed
approval_rule = ApprovalRule(
    "rule-test",
    name="My rule!",
    enabled=True,
    target_action_types=[TargetActionType.UPDATE],
    environment_identifier=EnvironmentIdentifierArgs(
        organization=environment.organization,
        project=environment.project,
        name=environment.name,
    ),
    approval_rule_config=ApprovalRuleConfigArgs(
        num_approvals_required=3,
        allow_self_approval=True,
        require_reapproval_on_change=True,
        eligible_approvers=[
            EligibleApproverArgs(
                rbac_permission=RbacPermission.WRITE,
            ),
            EligibleApproverArgs(
                user="pulumi-bot",
            ),
        ]
    )
)
