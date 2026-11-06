// Copyright 2026, Pulumi Corporation.

//go:build nodejs || python || go || dotnet || java || all

package examples

import (
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
	runSmokeTestAfterInstall(t, lang, nil, opts...)
}

// runSmokeTestAfterInstall is runSmokeTest with a hook that runs after the
// program is copied and installed and before `pulumi up`.
func runSmokeTestAfterInstall(
	t *testing.T,
	lang string,
	afterInstall func(t *testing.T, test *pulumitest.PulumiTest),
	opts ...opttest.Option,
) {
	digits := generateRandomFiveDigits()
	test := newExampleTest(t, lang+"-smoke", opts...)
	setConfig(t, test, map[string]string{"organizationName": getOrgName(), "digits": digits})
	if afterInstall != nil {
		afterInstall(t, test)
	}

	up := runPulumiTest(t, test)

	assert.Equal(t, "test-smoke-"+lang+"-"+digits, up.Outputs["environmentName"].Value,
		"getEnvironment should return the environment the program created")
	token := up.Outputs["tokenValue"]
	assert.True(t, token.Secret, "the token value should be a secret output")
	assert.NotEmpty(t, token.Value, "the token value should not be empty")
}
