// Copyright 2026, Pulumi Corporation.

//go:build go || dotnet || java || all

package examples

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// runSmokeTest runs one of the <lang>-smoke programs. Each creates an
// environment from a string asset, an org access token whose value is a secret
// output, and reads the environment back with the getEnvironment invoke: the
// shapes that each language SDK handles in its own code.
func runSmokeTest(t *testing.T, lang string, opts ...opttest.Option) {
	digits := generateRandomFiveDigits()
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), lang+"-smoke"),
		append([]opttest.Option{
			inMemoryProvider(),
			opttest.UseAmbientBackend(),
			opttest.StackName(randomStackName()),
		}, opts...)...,
	)
	test.SetConfig(t, "organizationName", getOrgName())
	test.SetConfig(t, "digits", digits)

	up := runPulumiTest(t, test)

	assert.Equal(t, "test-smoke-"+lang+"-"+digits, up.Outputs["environmentName"].Value,
		"getEnvironment should return the environment the program created")
	token := up.Outputs["tokenValue"]
	assert.True(t, token.Secret, "the token value should be a secret output")
	assert.NotEmpty(t, token.Value, "the token value should not be empty")
}
