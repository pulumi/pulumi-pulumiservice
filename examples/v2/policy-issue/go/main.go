package main

import (
	v2 "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		serviceOrg := cfg.Get("serviceOrg")
		if serviceOrg == "" {
			serviceOrg = "service-provider-test-org"
		}
		issueId := cfg.Get("issueId")
		if issueId == "" {
			issueId = "example-issue-id"
		}

		_, err := v2.NewPolicyIssue(ctx, "issue", &v2.PolicyIssueArgs{
			OrgName:    pulumi.String(serviceOrg),
			IssueId:    pulumi.String(issueId),
			Priority:   pulumi.String("high"),
			Status:     pulumi.String("in_progress"),
			AssignedTo: pulumi.String("pulumi-bot"),
		})
		return err
	})
}
