//go:build yaml

package examples

import (
	"os"
	"path"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/engine"
	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"gopkg.in/yaml.v2"
)

func TestYamlTeamsExample(t *testing.T) {
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
}

func TestYamlStackTagsExample(t *testing.T) {

	// Set up tmpdir with a Pulumi.yml with no resources
	// mimicking the deletion of resource
	tmpdir := t.TempDir()
	newProgram := map[string]string{
		"name":        "yaml-stack-tags-example",
		"runtime":     "yaml",
		"description": "A minimal Pulumi YAML program",
	}

	b, err := yaml.Marshal(newProgram)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(path.Join(tmpdir, "Pulumi.yml"), b, 0666)
	if err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Quick:       true,
		SkipRefresh: true,
		Dir:         path.Join(cwd, ".", "yaml-stack-tags"),
		StackName:   "test-stack",
		PrepareProject: func(_ *engine.Projinfo) error {
			return nil
		},
		EditDirs: []integration.EditDir{
			{
				Dir: tmpdir,
			},
		},
	})
}
