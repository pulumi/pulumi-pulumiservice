// Go variant of canonical/04-deployment-pipeline.
// A git-driven Pulumi Deployments pipeline: source + executor settings,
// drift detection on a cron, TTL on ephemeral stacks. Behavioral twin of
// the sibling YAML program.

package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/stacks/deployments"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		org := stringOr(cfg.Get("organizationName"), "service-provider-test-org")
		project := stringOr(cfg.Get("project"), "infrastructure")
		stack := stringOr(cfg.Get("stack"), "production")

		_, err := deployments.NewSettings(ctx, "deploymentSettings", &deployments.SettingsArgs{
			Organization: pulumi.String(org),
			Project:      pulumi.String(project),
			Stack:        pulumi.String(stack),
			SourceContext: pulumi.ToMap(map[string]interface{}{
				"git": map[string]interface{}{
					"repoUrl": "https://github.com/acme-corp/infrastructure.git",
					"branch":  "refs/heads/main",
					"repoDir": "stacks/production",
				},
			}),
			Github: pulumi.ToMap(map[string]interface{}{
				"repository":          "acme-corp/infrastructure",
				"deployCommits":       true,
				"previewPullRequests": true,
				"pullRequestTemplate": true,
			}),
			ExecutorContext: pulumi.ToMap(map[string]interface{}{
				"executorImage": map[string]interface{}{
					"image": "pulumi/pulumi:latest",
				},
			}),
			OperationContext: pulumi.ToMap(map[string]interface{}{
				"preRunCommands":       []string{"npm ci"},
				"environmentVariables": map[string]string{"NODE_ENV": "production"},
			}),
		})
		if err != nil {
			return err
		}

		_, err = deployments.NewDriftSchedule(ctx, "driftCheck", &deployments.DriftScheduleArgs{
			Organization:  pulumi.String(org),
			Project:       pulumi.String(project),
			Stack:         pulumi.String(stack),
			ScheduleCron:  pulumi.String("0 */6 * * *"),
			AutoRemediate: pulumi.Bool(false),
		})
		if err != nil {
			return err
		}

		_, err = deployments.NewTtlSchedule(ctx, "ephemeralTtl", &deployments.TtlScheduleArgs{
			Organization:       pulumi.String(org),
			Project:            pulumi.String(project),
			Stack:              pulumi.String(stack),
			Timestamp:          pulumi.String("2026-12-31T00:00:00Z"),
			DeleteAfterDestroy: pulumi.Bool(true),
		})
		return err
	})
}

func stringOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
