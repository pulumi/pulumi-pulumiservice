//go:build yaml

package examples

import (
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/pulumi/pulumi/pkg/v3/engine"
	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/stretchr/testify/assert"
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
			Quick:       true,
			SkipRefresh: true,
			Dir:         path.Join(cwd, ".", "yaml-teams"),
			// don't prepare project at all, not required for yaml
			PrepareProject: func(_ *engine.Projinfo) error {
				return nil
			},
		})
	})

}

func TestYamlStackTagsExample(t *testing.T) {

	// Set up tmpdir with a Pulumi.yml with no resources
	// mimicking the deletion of resource
	newProgram := YamlProgram{
		Name:        "yaml-stack-tags-example",
		Runtime:     "yaml",
		Description: "A minimal Pulumi YAML program",
	}

	tmpdir := writePulumiYaml(t, newProgram)

	cwd, _ := os.Getwd()

	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick:       true,
		SkipRefresh: true,
		Dir:         path.Join(cwd, ".", "yaml-stack-tags"),
		StackName:   fmt.Sprintf("%s/%s", ServiceProviderTestOrg, "test-stack"),
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

func TestYamlTeamStackPermissionsExample(t *testing.T) {
	cwd, _ := os.Getwd()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick:       true,
		SkipRefresh: true,
		// Name is specified in yaml-team-stack-permissions/Pulumi.yaml, so this has to be consistent
		StackName: fmt.Sprintf("%s/%s", ServiceProviderTestOrg, "dev"),
		Dir:       path.Join(cwd, ".", "yaml-team-stack-permissions"),
		PrepareProject: func(_ *engine.Projinfo) error {
			return nil
		},
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
