//go:build python || all
// +build python all

package examples

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
)

func TestPythonSmokeExample(t *testing.T) {
	// opttest.PythonLink installs into the system interpreter, not the
	// program's virtualenv, so install the local SDK into the virtualenv that
	// `pulumi install` created.
	runSmokeTest(t, "py", func(t *testing.T, test *pulumitest.PulumiTest) {
		sdk := filepath.Join(getCwd(t), "..", "sdk", "python", "bin")
		python := filepath.Join(test.WorkingDir(), "venv", "bin", "python")
		cmd := exec.Command(python, "-m", "pip", "install", "-e", sdk)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("installing %s into the program's virtualenv: %v\n%s", sdk, err, out)
		}
	})
}
