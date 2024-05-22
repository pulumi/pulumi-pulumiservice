//go:build dotnet || all

package examples

import (
	"path"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestDotnetTeamsExamples(t *testing.T) {
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir:         path.Join(getCwd(t), "cs-teams"),
		SkipRefresh: true,
		Dependencies: []string{
			"Pulumi.PulumiService",
		},
	})
}

func TestDotnetSchedulesExamples(t *testing.T) {
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir:         path.Join(getCwd(t), "cs-schedules"),
		StackName:   "test-stack-" + digits,
		SkipRefresh: true,
		Config: map[string]string{
			"digits": digits,
		},
		Dependencies: []string{
			"Pulumi.PulumiService",
		},
	})
}

func TestDotnetEnvironmentsExamples(t *testing.T) {
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir:         path.Join(getCwd(t), "cs-environments"),
		SkipRefresh: true,
		Dependencies: []string{
			"Pulumi.PulumiService",
		},
	})
}
