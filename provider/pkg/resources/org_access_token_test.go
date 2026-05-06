package resources

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

// State written by the pre-infer OrgAccessToken implementation embedded a
// duplicate `__inputs` map alongside the real outputs. After migrating to
// infer, decoding such state fails with "Unrecognized field '__inputs'" unless
// migrateOrgAccessTokenLegacyInputs runs first.
func TestOrgAccessTokenLegacyInputsMigration(t *testing.T) {
	t.Run("migrates legacy state", func(t *testing.T) {
		legacyInputs := property.NewMap(map[string]property.Value{
			"name":             property.New("admin-token"),
			"organizationName": property.New("my-org"),
			"description":      property.New("example org token"),
			"admin":            property.New(true),
		})
		legacy := property.NewMap(map[string]property.Value{
			"__inputs":         property.New(legacyInputs),
			"name":             property.New("admin-token"),
			"organizationName": property.New("my-org"),
			"description":      property.New("example org token"),
			"admin":            property.New(true),
			"value":            property.New("tok-secret-value"),
		})

		got, err := migrateOrgAccessTokenLegacyInputs(t.Context(), legacy)
		require.NoError(t, err)
		desc := "example org token"
		admin := true
		assert.Equal(t, &OrgAccessTokenState{
			OrgAccessTokenInput: OrgAccessTokenInput{
				Name:             "admin-token",
				OrganizationName: "my-org",
				Description:      &desc,
				Admin:            &admin,
			},
			Value: "tok-secret-value",
		}, got.Result)
	})

	t.Run("migrates legacy state with optional fields omitted", func(t *testing.T) {
		legacy := property.NewMap(map[string]property.Value{
			"__inputs":         property.New(property.NewMap(nil)),
			"name":             property.New("plain-token"),
			"organizationName": property.New("my-org"),
			"value":            property.New("tok-secret-value"),
		})

		got, err := migrateOrgAccessTokenLegacyInputs(t.Context(), legacy)
		require.NoError(t, err)
		assert.Equal(t, &OrgAccessTokenState{
			OrgAccessTokenInput: OrgAccessTokenInput{
				Name:             "plain-token",
				OrganizationName: "my-org",
			},
			Value: "tok-secret-value",
		}, got.Result)
	})

	t.Run("no-op for already-migrated state", func(t *testing.T) {
		current := property.NewMap(map[string]property.Value{
			"name":             property.New("admin-token"),
			"organizationName": property.New("my-org"),
			"value":            property.New("tok-secret-value"),
		})

		got, err := migrateOrgAccessTokenLegacyInputs(t.Context(), current)
		require.NoError(t, err)
		assert.Nil(t, got.Result, "migration must not fire when __inputs is absent")
	})

	t.Run("registered on the resource", func(t *testing.T) {
		migrations := (&OrgAccessToken{}).StateMigrations(t.Context())
		assert.Len(t, migrations, 1)
	})
}

func TestSplitOrgAccessTokenId(t *testing.T) {
	t.Run("Splits org access token id", func(t *testing.T) {
		tokenID := "org/name/id"

		org, name, id, err := splitOrgAccessTokenID(tokenID)
		assert.NoError(t, err)

		assert.Equal(t, "org", org)
		assert.Equal(t, "name", name)
		assert.Equal(t, "id", id)
	})

	t.Run("Splits org access token id with name with slashes", func(t *testing.T) {
		tokenID := "org/name/with/slashes/id" //nolint:gosec // test data, not a real credential

		org, name, id, err := splitOrgAccessTokenID(tokenID)
		assert.NoError(t, err)

		assert.Equal(t, "org", org)
		assert.Equal(t, "name/with/slashes", name)
		assert.Equal(t, "id", id)
	})

	t.Run("Splits org access token id with invalid id", func(t *testing.T) {
		tokenID := "org/badname"

		_, _, _, err := splitOrgAccessTokenID(tokenID)
		assert.ErrorContains(t, err, fmt.Sprintf("%q is invalid, must be of the form organization/name/tokenId", tokenID))
	})
}
