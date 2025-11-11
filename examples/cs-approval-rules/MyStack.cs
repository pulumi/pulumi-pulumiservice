using Pulumi;
using Pulumi.PulumiService;
using Pulumi.PulumiService.Inputs;
using System;
using System.Collections.Generic;

class MyStack : Pulumi.Stack
{
    public MyStack()
    {
        var config = new Pulumi.Config();
        String yaml = """
        values:
          myKey1: "myValue1"
        """;

        // Create an environment that will be used as the target for our approval rule
        var environment = new Pulumi.PulumiService.Environment(
            "testing-environment",
            new EnvironmentArgs {
                Organization = "service-provider-test-org",
                Project = "test-project",
                Name = "testing-environment-approval-cs-" + config.Require("digits"),
                Yaml = new StringAsset(yaml)
            }
        );

        // Create an approval rule that governs who can approve updates to the environment
        // This rule requires 3 approvals from eligible approvers before any update can proceed
        var approvalRule = new Pulumi.PulumiService.ApprovalRule(
            "rule-test",
            new ApprovalRuleArgs {
                Name = "My rule!",
                Enabled = true,
                TargetActionTypes = new List<Pulumi.PulumiService.TargetActionType> { TargetActionType.Update },
                EnvironmentIdentifier = new EnvironmentIdentifierArgs {
                    Organization = environment.Organization,
                    Project = environment.Project,
                    Name = environment.Name
                },
                ApprovalRuleConfig = new ApprovalRuleConfigArgs {
                    NumApprovalsRequired = 3,
                    AllowSelfApproval = true,
                    RequireReapprovalOnChange = true,
                    EligibleApprovers = new List<EligibleApproverArgs> {
                        new EligibleApproverArgs {
                            RbacPermission = RbacPermission.Write
                        },
                        new EligibleApproverArgs {
                            User = "pulumi-bot"
                        },
                    }
                }
            }
        );
    }
}
