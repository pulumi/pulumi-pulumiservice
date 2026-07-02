// Copyright 2016-2026, Pulumi Corporation.

package rest

// Shared string constants used across the rest package tests. They exist to
// satisfy goconst and to give repeated fixture values a single source of truth.
// (descriptionKey lives in spec.go since it is also used by non-test code.)
const (
	// Map keys / field names.
	nameKey             = "name"
	orgKey              = "org"
	orgNameKey          = "orgName"
	modeKey             = "mode"
	accountsKey         = "accounts"
	policyGroupKey      = "policyGroup"
	tokenValueKey       = "tokenValue"
	valueKey            = "value"
	organizationNameVal = "organizationName"
	widgetIDVal         = "widgetID"
	outputVal           = "output"

	// Fixture values.
	grpVal       = "grp"
	stacksVal    = "stacks"
	devVal       = "dev"
	myProjectVal = "my-project"
	acct1Val     = "acct-1"
	acmeVal      = "acme"
	infraVal     = "infra"
	infraTeamVal = "infra team"
	newVal       = "new"
	originalVal  = "original"
	rotatedVal   = "rotated"
	goneVal      = "gone"

	// Composite identities.
	devProjectID = "test-org/grp/dev/my-project"
	acct1ID      = "test-org/grp/acct-1"
	acmeGoneID   = "acme/gone"

	// Operation names and id/path formats.
	createThingOp = "CreateThing"
	putThingOp    = "PutThing"
	orgIDFormat   = "{org}/{id}"
	orgFormat     = "{org}"

	// Mock request keys and response bodies.
	postThingsAcme       = "POST /things/acme"
	getThingsAcme        = "GET /things/acme"
	putThingsAcme        = "PUT /things/acme"
	deleteThingsAcmeGone = "DELETE /things/acme/gone"
	unexpectedBody       = "unexpected"
	notFoundBody         = `{"error":"not found"}`
	orgAcmeNewBody       = `{"org":"acme","value":"new"}`

	// Timestamps.
	createdTimestamp = "2026-05-05T00:00:00Z"

	// Additional shared fixture literals.
	testOrgName          = "test-org"
	routingProjectKey    = "routingProject"
	orgProjectNameFormat = "{org}/{project}/{name}"
	getThingOp           = "GetThing"
	fooVal               = "foo"
	thing1ID             = "thing-1"
	createdKey           = "created"
)
