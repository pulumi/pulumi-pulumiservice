package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"

	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		conf := config.New(ctx, "")
		yaml := `values:
  myKey1: "myValue1"`

		// Create an environment that will be used as the target for our approval rule
		environment, err := pulumiservice.NewEnvironment(ctx, "testing-environment", &pulumiservice.EnvironmentArgs{
			Name:         pulumi.String("testing-environment-approval-go-" + conf.Require("digits")),
			Project:      pulumi.String("test-project"),
			Organization: pulumi.String("service-provider-test-org"),
			Yaml:         pulumi.NewStringAsset(yaml),
		})
		if err != nil {
			return err
		}

		// Create an approval rule that governs who can approve updates to the environment
		// This rule requires 3 approvals from eligible approvers before any update can proceed
		_, err = pulumiservice.NewApprovalRule(ctx, "rule-test", &pulumiservice.ApprovalRuleArgs{
			Name:              pulumi.String("My rule!"),
			Enabled:           pulumi.Bool(true),
			TargetActionTypes: pulumiservice.TargetActionTypeArray{pulumiservice.TargetActionTypeUpdate},
			EnvironmentIdentifier: &pulumiservice.EnvironmentIdentifierArgs{
				Organization: environment.Organization,
				Project:      environment.Project,
				Name:         environment.Name,
			},
			ApprovalRuleConfig: &pulumiservice.ApprovalRuleConfigArgs{
				NumApprovalsRequired:      pulumi.Int(3),
				AllowSelfApproval:         pulumi.Bool(true),
				RequireReapprovalOnChange: pulumi.Bool(true),
				EligibleApprovers: pulumiservice.EligibleApproverArray{
					&pulumiservice.EligibleApproverArgs{
						RbacPermission: pulumiservice.RbacPermissionWrite,
					},
					&pulumiservice.EligibleApproverArgs{
						User: pulumi.String("pulumi-bot"),
					},
					&pulumiservice.EligibleApproverArgs{
						User: pulumi.String("IaroslavTitov"),
					},
				},
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
}
