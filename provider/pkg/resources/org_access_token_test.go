package resources

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

const (
	gcTokSecretValue = "tok-secret-value"
	gcMyOrg          = "my-org"
)

// State written by the pre-infer OrgAccessToken implementation embedded a
// duplicate `__inputs` map alongside the real outputs. After migrating to
// infer, decoding such state fails with "Unrecognized field '__inputs'" unless
// migrateOrgAccessTokenLegacyInputs runs first.
func TestOrgAccessTokenLegacyInputsMigration(t *testing.T) {
	t.Run("migrates legacy state", func(t *testing.T) {
		legacyInputs := property.NewMap(map[string]property.Value{
			gcName:             property.New("admin-token"),
			gcOrganizationName: property.New(gcMyOrg),
			gcDescription:      property.New("example org token"),
			gcAdmin:            property.New(true),
		})
		legacy := property.NewMap(map[string]property.Value{
			gcInputs:           property.New(legacyInputs),
			gcName:             property.New("admin-token"),
			gcOrganizationName: property.New(gcMyOrg),
			gcDescription:      property.New("example org token"),
			gcAdmin:            property.New(true),
			gcValue:            property.New(gcTokSecretValue),
		})

		got, err := migrateOrgAccessTokenLegacyInputs(t.Context(), legacy)
		require.NoError(t, err)
		desc := "example org token"
		admin := true
		assert.Equal(t, &OrgAccessTokenState{
			OrgAccessTokenInput: OrgAccessTokenInput{
				Name:             "admin-token",
				OrganizationName: gcMyOrg,
				Description:      &desc,
				Admin:            &admin,
			},
			Value: gcTokSecretValue,
		}, got.Result)
	})

	t.Run("migrates legacy state with optional fields omitted", func(t *testing.T) {
		legacy := property.NewMap(map[string]property.Value{
			gcInputs:           property.New(property.NewMap(nil)),
			gcName:             property.New("plain-token"),
			gcOrganizationName: property.New(gcMyOrg),
			gcValue:            property.New(gcTokSecretValue),
		})

		got, err := migrateOrgAccessTokenLegacyInputs(t.Context(), legacy)
		require.NoError(t, err)
		assert.Equal(t, &OrgAccessTokenState{
			OrgAccessTokenInput: OrgAccessTokenInput{
				Name:             "plain-token",
				OrganizationName: gcMyOrg,
			},
			Value: gcTokSecretValue,
		}, got.Result)
	})

	t.Run("no-op for already-migrated state", func(t *testing.T) {
		current := property.NewMap(map[string]property.Value{
			gcName:             property.New("admin-token"),
			gcOrganizationName: property.New(gcMyOrg),
			gcValue:            property.New(gcTokSecretValue),
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

		assert.Equal(t, gcOrg, org)
		assert.Equal(t, gcName, name)
		assert.Equal(t, "id", id)
	})

	t.Run("Splits org access token id with name with slashes", func(t *testing.T) {
		tokenID := "org/name/with/slashes/id" //nolint:gosec // test data, not a real credential

		org, name, id, err := splitOrgAccessTokenID(tokenID)
		assert.NoError(t, err)

		assert.Equal(t, gcOrg, org)
		assert.Equal(t, "name/with/slashes", name)
		assert.Equal(t, "id", id)
	})

	t.Run("Splits org access token id with invalid id", func(t *testing.T) {
		tokenID := "org/badname"

		_, _, _, err := splitOrgAccessTokenID(tokenID)
		assert.ErrorContains(t, err, fmt.Sprintf("%q is invalid, must be of the form organization/name/tokenId", tokenID))
	})
}

type orgAccessTokenClientMock struct {
	config.Client
	create func(
		ctx context.Context, name, orgName, description string, admin bool, expires int64, roleID string,
	) (*pulumiapi.AccessToken, error)
	get func(ctx context.Context, tokenID, orgName string) (*pulumiapi.AccessToken, error)
}

func (c *orgAccessTokenClientMock) CreateOrgAccessToken(
	ctx context.Context, name, orgName, description string, admin bool, expires int64, roleID string,
) (*pulumiapi.AccessToken, error) {
	return c.create(ctx, name, orgName, description, admin, expires, roleID)
}

func (c *orgAccessTokenClientMock) GetOrgAccessToken(
	ctx context.Context, tokenID, orgName string,
) (*pulumiapi.AccessToken, error) {
	return c.get(ctx, tokenID, orgName)
}

const (
	// 2027-06-30T17:00:00Z
	testExpiryEpoch int64 = 1814374800
	// The same instant as testExpiryEpoch, written with an offset.
	testExpiryWithOffset = "2027-06-30T12:00:00-05:00"
)

func TestParseTokenExpiry(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want int64
	}{
		{"2027-06-30T17:00:00Z", testExpiryEpoch},
		{testExpiryWithOffset, testExpiryEpoch},
		{"2027-06-30", testExpiryEpoch - 17*60*60},
	} {
		got, err := parseTokenExpiry(tc.in)
		require.NoError(t, err, tc.in)
		assert.Equal(t, tc.want, got.Unix(), tc.in)
	}

	for _, in := range []string{"", "30d", "2027-06-30 17:00:00", "06/30/2027", "2027-13-01"} {
		_, err := parseTokenExpiry(in)
		assert.ErrorContains(t, err, "is not a valid expiry", in)
	}
}

func TestFormatTokenExpiry(t *testing.T) {
	assert.Equal(t, "2027-06-30T17:00:00Z", formatTokenExpiry(testExpiryEpoch))
	assert.Equal(t, "2027-06-30", formatTokenExpiry(testExpiryEpoch-17*60*60))
}

func TestReadTokenExpiry(t *testing.T) {
	ptr := func(s string) *string { return &s }

	assert.Nil(t, readTokenExpiry(0, ptr("2027-06-30T17:00:00Z")))
	// Same instant, different spelling: keep the user's value.
	assert.Equal(t, testExpiryWithOffset, *readTokenExpiry(testExpiryEpoch, ptr(testExpiryWithOffset)))
	// Different instant (changed outside Pulumi): report the API's value.
	assert.Equal(t, "2027-06-30T17:00:00Z", *readTokenExpiry(testExpiryEpoch, ptr("2027-01-01")))
	// Import: no prior value.
	assert.Equal(t, "2027-06-30T17:00:00Z", *readTokenExpiry(testExpiryEpoch, nil))
}

func TestOrgAccessTokenCheck(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	orig := tokenExpiryNow
	tokenExpiryNow = func() time.Time { return now }
	t.Cleanup(func() { tokenExpiryNow = orig })

	inputs := func(extra map[string]property.Value) property.Map {
		m := map[string]property.Value{
			gcName:             property.New("ci-token"),
			gcOrganizationName: property.New(gcMyOrg),
		}
		for k, v := range extra {
			m[k] = v
		}
		return property.NewMap(m)
	}
	check := func(newInputs, oldInputs property.Map) []string {
		resp, err := (&OrgAccessToken{}).Check(t.Context(), infer.CheckRequest{
			NewInputs: newInputs,
			OldInputs: oldInputs,
		})
		require.NoError(t, err)
		var reasons []string
		for _, f := range resp.Failures {
			reasons = append(reasons, f.Reason)
		}
		return reasons
	}

	t.Run("valid date-time", func(t *testing.T) {
		assert.Empty(t, check(inputs(map[string]property.Value{
			gcExpires: property.New("2027-06-30T17:00:00Z"),
		}), property.Map{}))
	})

	t.Run("unparseable", func(t *testing.T) {
		reasons := check(inputs(map[string]property.Value{gcExpires: property.New("30d")}), property.Map{})
		require.Len(t, reasons, 1)
		assert.Contains(t, reasons[0], "is not a valid expiry")
	})

	t.Run("in the past", func(t *testing.T) {
		reasons := check(inputs(map[string]property.Value{
			gcExpires: property.New("2026-10-15T11:00:00Z"),
		}), property.Map{})
		require.Len(t, reasons, 1)
		assert.Contains(t, reasons[0], "is in the past")
	})

	t.Run("more than two years away", func(t *testing.T) {
		reasons := check(inputs(map[string]property.Value{
			gcExpires: property.New("2028-10-16T00:00:00Z"),
		}), property.Map{})
		require.Len(t, reasons, 1)
		assert.Contains(t, reasons[0], "more than 2 years away")
	})

	t.Run("unchanged expired value passes", func(t *testing.T) {
		in := inputs(map[string]property.Value{gcExpires: property.New("2026-01-01T00:00:00Z")})
		assert.Empty(t, check(in, in))
	})

	t.Run("changed to an expired value fails", func(t *testing.T) {
		old := inputs(map[string]property.Value{gcExpires: property.New("2027-01-01T00:00:00Z")})
		in := inputs(map[string]property.Value{gcExpires: property.New("2026-01-01T00:00:00Z")})
		assert.Len(t, check(in, old), 1)
	})

	t.Run("unknown expiry is skipped", func(t *testing.T) {
		assert.Empty(t, check(inputs(map[string]property.Value{
			gcExpires: property.New(property.Computed),
		}), property.Map{}))
	})

	t.Run("admin and roleId together", func(t *testing.T) {
		reasons := check(inputs(map[string]property.Value{
			gcAdmin:     property.New(true),
			gcRoleIDKey: property.New(gcRoleID),
		}), property.Map{})
		require.Len(t, reasons, 1)
		assert.Contains(t, reasons[0], "roleId can't be combined with admin: true")
	})

	t.Run("roleId alone", func(t *testing.T) {
		assert.Empty(t, check(inputs(map[string]property.Value{
			gcRoleIDKey: property.New(gcRoleID),
		}), property.Map{}))
	})
}

func TestOrgAccessTokenCreateSendsExpiryAndRole(t *testing.T) {
	var gotExpires int64
	var gotRoleID string
	mock := &orgAccessTokenClientMock{
		create: func(
			_ context.Context, _, _, _ string, _ bool, expires int64, roleID string,
		) (*pulumiapi.AccessToken, error) {
			gotExpires, gotRoleID = expires, roleID
			return &pulumiapi.AccessToken{ID: "tok-id", TokenValue: gcTokSecretValue}, nil
		},
	}
	ctx := config.WithMockClient(t.Context(), mock)

	expires := testExpiryWithOffset
	roleID := gcRoleID
	resp, err := (&OrgAccessToken{}).Create(ctx, infer.CreateRequest[OrgAccessTokenInput]{
		Inputs: OrgAccessTokenInput{
			Name:             "ci-token",
			OrganizationName: gcMyOrg,
			Expires:          &expires,
			RoleID:           &roleID,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, testExpiryEpoch, gotExpires)
	assert.Equal(t, roleID, gotRoleID)
	assert.Equal(t, expires, *resp.Output.Expires)
}

func TestOrgAccessTokenReadKeepsExpirySpelling(t *testing.T) {
	mock := &orgAccessTokenClientMock{
		get: func(_ context.Context, _, _ string) (*pulumiapi.AccessToken, error) {
			return &pulumiapi.AccessToken{
				ID:      "tok-id",
				Name:    "ci-token",
				Expires: testExpiryEpoch,
				RoleID:  gcRoleID,
			}, nil
		},
	}
	ctx := config.WithMockClient(t.Context(), mock)

	expires := testExpiryWithOffset
	resp, err := (&OrgAccessToken{}).Read(ctx, infer.ReadRequest[OrgAccessTokenInput, OrgAccessTokenState]{
		ID: orgAccessTokenID(gcMyOrg, "ci-token", "tok-id"),
		State: OrgAccessTokenState{
			OrgAccessTokenInput: OrgAccessTokenInput{Expires: &expires},
			Value:               gcTokSecretValue,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Inputs.Expires)
	assert.Equal(t, expires, *resp.Inputs.Expires)
	require.NotNil(t, resp.Inputs.RoleID)
	assert.Equal(t, gcRoleID, *resp.Inputs.RoleID)
	assert.Equal(t, gcTokSecretValue, resp.State.Value)
}
