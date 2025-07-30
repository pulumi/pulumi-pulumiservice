import * as service from "@pulumi/pulumiservice";
import * as pulumi from "@pulumi/pulumi";

let config = new pulumi.Config();
const digits = config.require("digits");

// Create an environment that will be used as the target for our approval rule
var environment = new service.Environment("testing-environment", {
  organization: "service-provider-test-org",
  project: "test-project",
  name: `testing-environment-approval-ts-${digits}`,
  yaml: new pulumi.asset.StringAsset(
`values:
  myKey1: "myValue1"`
  )
})

// Create an approval rule that governs who can approve updates to the environment
// This rule requires 3 approvals from eligible approvers before any update can proceed
var approvalRule = new service.ApprovalRule("rule-test", {
  name: "My rule!",
  enabled: true,
  targetActionType: "update",
  environmentIdentifier: {
    organization: environment.organization,
    project: environment.project,
    name: environment.name,
  },
  approvalRuleConfig: {
    numApprovalsRequired: 3,
    allowSelfApproval: true,
    requireReapprovalOnChange: true,
    eligibleApprovers: [
      {
        rbacPermission: "environment:write",
      },
      {
        user: "pulumi-bot",
      },
      {
        user: "IaroslavTitov",
      },
    ]
  }
})
