//go:build yaml

package examples

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/pulumi/pulumi/pkg/v3/engine"
	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type Resource struct {
	Type       string                 `yaml:"type"`
	Properties map[string]interface{} `yaml:"properties"`
}
type YamlProgram struct {
	Name        string              `yaml:"name"`
	Runtime     string              `yaml:"runtime"`
	Description string              `yaml:"description"`
	Resources   map[string]Resource `yaml:"resources"`
}

const (
	ServiceProviderTestOrg = "service-provider-test-org"
)

func TestYamlTeamsExample(t *testing.T) {

	// This test builds a repro of https://github.com/pulumi/pulumi-pulumiservice/issues/73.
	// The scenario is where a team is partially created due to a invalid member being specified.
	// This leaves a team resource around even though the creation failed.
	// A subsequent update to the stack with the invalid team member removed
	// should run Update against the partially created team, instead of attempting
	// to Create the team
	t.Run("Properly Updates After Partial Creation Failure", func(t *testing.T) {
		teamName := uuid.NewString()
		projectName := "yaml-teams-fail-" + uuid.NewString()[0:10]

		// this function writes out a Pulumi.yaml file to a temp directory
		writeTeamProgram := func(members []string) string {
			prog := YamlProgram{
				Name:    projectName,
				Runtime: "yaml",
				Resources: map[string]Resource{
					"team": {
						Type: "pulumiservice:index:Team",
						Properties: map[string]interface{}{
							"name":             teamName,
							"teamType":         "pulumi",
							"displayName":      teamName,
							"organizationName": ServiceProviderTestOrg,
							"members":          members,
						},
					},
				},
			}

			return writePulumiYaml(t, prog)
		}

		pulumiBot := "pulumi-bot"
		providerUser := "service-provider-example-user"

		notFoundDir := writeTeamProgram([]string{pulumiBot, providerUser, "not-found-user"})
		correctUpdateDir := writeTeamProgram([]string{pulumiBot, providerUser})

		first := &strings.Builder{}
		firstOut := io.MultiWriter(os.Stdout, first)

		second := &strings.Builder{}
		secondOut := io.MultiWriter(os.Stdout, second)

		integration.ProgramTest(t, &integration.ProgramTestOptions{
			Quick:         true,
			ExpectFailure: true,
			SkipRefresh:   true,
			Dir:           notFoundDir,
			Stdout:        firstOut,
			Stderr:        firstOut,
			// don't prepare project at all, not required for yaml
			PrepareProject: func(_ *engine.Projinfo) error {
				return nil
			},
			EditDirs: []integration.EditDir{
				{
					Dir: correctUpdateDir,
					// Additive specifies that we're copying our directory on top of the previous one.
					// This overwrites the previous Pulumi.yaml.
					Additive: true,
					Verbose:  true,
					Stdout:   secondOut,
					Stderr:   secondOut,
					// explicitly do not expect a failure
					ExpectFailure: false,
				},
			},
		})

		// ensure first run's output printed an error that user was not found
		assert.Contains(t, first.String(), "User 'not-found-user' not found")

		// ensure second run's output showed that members field was updated
		assert.Contains(t, second.String(), "[diff: ~members]")
	})

	t.Run("Yaml Teams Example", func(t *testing.T) {
		cwd, _ := os.Getwd()
		integration.ProgramTest(t, &integration.ProgramTestOptions{
			Quick: true,
			Dir:   path.Join(cwd, ".", "yaml-teams"),
			// don't prepare project at all, not required for yaml
			PrepareProject: func(_ *engine.Projinfo) error {
				return nil
			},
		})
	})

}

func TestYamlStackTagsExample(t *testing.T) {

	// yaml-stack-tags-example applies tags to it's own stack. To do this, we need to
	// first create an empty stack, then add the stack tag.

	tmpdir := writePulumiYaml(t, YamlProgram{
		Name:        "yaml-stack-tags-example",
		Runtime:     "yaml",
		Description: "A minimal Pulumi YAML program",
	})

	cwd, err := os.Getwd()
	require.NoError(t, err)

	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick: true,
		Dir:   tmpdir,
		EditDirs: []integration.EditDir{
			{
				Dir: path.Join(cwd, ".", "yaml-stack-tags"),
			},
			// Reapply the same thing again, except this time we expect there to be no changes
			{
				Dir:             path.Join(cwd, ".", "yaml-stack-tags"),
				ExpectNoChanges: true,
			},
		},

		// Setting Config and PrepareProject works around the bug introduced in
		// https://github.com/pulumi/pulumi/pull/14695:
		//
		// - Setting a config value ensures that the stack file is created.
		//
		// - PrepareProject is used to force the bookkeeping file to be created.
		//
		// Once https://github.com/pulumi/pulumi/pull/15863 merges, both fields
		// can be removed.
		Config: map[string]string{"ensure-config": "true"},
		PrepareProject: func(info *engine.Projinfo) error {
			return os.MkdirAll(filepath.Join(info.Root, ".pulumi"), 0700)
		},
	})
}

func TestYamlDeploymentSettingsExample(t *testing.T) {

	// Set up tmpdir with a Pulumi.yml with no resources
	// mimicking the deletion of resource
	newProgram := YamlProgram{
		Name:        "yaml-deployment-settings-example",
		Runtime:     "yaml",
		Description: "Deployment settings test",
	}

	tmpdir := writePulumiYaml(t, newProgram)

	cwd, _ := os.Getwd()
	digits := generateRandomFiveDigits()

	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick:     true,
		Dir:       path.Join(cwd, ".", "yaml-deployment-settings"),
		StackName: "test-stack-" + digits,
		Config: map[string]string{
			"digits": digits,
		},
		PrepareProject: func(_ *engine.Projinfo) error {
			return nil
		},
		EditDirs: []integration.EditDir{
			{
				Dir: tmpdir,
			},
			// Reapply the same thing again, except this time we expect there to be no changes
			{
				Dir:             tmpdir,
				ExpectNoChanges: true,
			},
		},
	})
}

func TestYamlDeploymentSettingsNoSourceExample(t *testing.T) {

	// Set up tmpdir with a Pulumi.yml with no resources
	// mimicking the deletion of resource
	newProgram := YamlProgram{
		Name:    "yaml-deployment-settings-example-no-source",
		Runtime: "yaml",
	}

	tmpdir := writePulumiYaml(t, newProgram)

	cwd, _ := os.Getwd()
	digits := generateRandomFiveDigits()

	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick:     true,
		Dir:       path.Join(cwd, ".", "yaml-deployment-settings-no-source"),
		StackName: "test-stack-" + digits,
		Config: map[string]string{
			"digits": digits,
		},
		PrepareProject: func(_ *engine.Projinfo) error {
			return nil
		},
		EditDirs: []integration.EditDir{
			{
				Dir: tmpdir,
			},
			// Reapply the same thing again, except this time we expect there to be no changes
			{
				Dir:             tmpdir,
				ExpectNoChanges: true,
			},
		},
	})
}

func TestYamlTeamAccessTokenExample(t *testing.T) {
	cwd, _ := os.Getwd()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick: true,
		Dir:   path.Join(cwd, ".", "yaml-team-token"),
		PrepareProject: func(_ *engine.Projinfo) error {
			return nil
		},
	})
}

func TestYamlOrgAccessTokenExample(t *testing.T) {
	cwd, _ := os.Getwd()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick: true,
		Dir:   path.Join(cwd, ".", "yaml-org-token"),
		PrepareProject: func(_ *engine.Projinfo) error {
			return nil
		},
	})
}

func TestYamlTeamStackPermissionsExample(t *testing.T) {
	cwd, _ := os.Getwd()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick: true,
		Dir:   path.Join(cwd, ".", "yaml-team-stack-permissions"),
	})
}

func TestYamlWebhookExample(t *testing.T) {
	cwd := getCwd(t)
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(cwd, ".", "yaml-webhooks"),
		PrepareProject: func(p *engine.Projinfo) error {
			return nil
		},
	})
}

func TestYamlSchedulesExample(t *testing.T) {
	cwd := getCwd(t)
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir:       path.Join(cwd, ".", "yaml-schedules"),
		StackName: "test-stack-" + digits,
		Config: map[string]string{
			"digits": digits,
		},
		PrepareProject: func(p *engine.Projinfo) error {
			return nil
		},
	})
}

func TestYamlEnvironmentsExample(t *testing.T) {
	cwd := getCwd(t)
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(cwd, ".", "yaml-environments"),
		Config: map[string]string{
			"digits": digits,
		},
		PrepareProject: func(p *engine.Projinfo) error {
			return nil
		},
	})
}

func TestYamlAgentPoolsExample(t *testing.T) {
	cwd := getCwd(t)
	digits := generateRandomFiveDigits()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(cwd, ".", "yaml-agent-pools"),
		Config: map[string]string{
			"digits": digits,
		},
		PrepareProject: func(p *engine.Projinfo) error {
			return nil
		},
	})
}

func TestYamlTemplateSourcesExample(t *testing.T) {
	cwd := getCwd(t)
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir: path.Join(cwd, ".", "yaml-template-sources"),
		PrepareProject: func(p *engine.Projinfo) error {
			return nil
		},
	})
}

func writePulumiYaml(t *testing.T, yamlContents interface{}) string {
	tmpdir := t.TempDir()
	b, err := yaml.Marshal(yamlContents)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(path.Join(tmpdir, "Pulumi.yaml"), b, 0666)
	if err != nil {
		t.Fatal(err)
	}
	return tmpdir
}
