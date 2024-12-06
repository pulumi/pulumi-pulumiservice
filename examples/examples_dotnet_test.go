//go:build dotnet || all
// +build dotnet all

package examples

import (
	"path"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestDotnetTeamsExamples(t *testing.T) {
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(getCwd(t), "cs-teams"),
		Dependencies: []string{
			"Pulumi.PulumiService",
		},
	})
}

func TestDotnetSchedulesExamples(t *testing.T) {
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir:       path.Join(getCwd(t), "cs-schedules"),
		StackName: "test-stack-" + digits,
		Config: map[string]string{
			"digits": digits,
		},
		Dependencies: []string{
			"Pulumi.PulumiService",
		},
	})
}

func TestDotnetEnvironmentsExamples(t *testing.T) {
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(getCwd(t), "cs-environments"),
		Config: map[string]string{
			"digits": digits,
		},
		Dependencies: []string{
			"Pulumi.PulumiService",
		},
	})
}
