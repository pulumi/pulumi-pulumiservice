// Copyright 2016-2026, Pulumi Corporation.

package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

func TestMemberInfoFromRoles(t *testing.T) {
	for display, slug := range map[string]string{
		"Admin":      "admin",
		"Builder":    "builder",
		"Maintainer": "maintainer",
		"Reader":     "reader",
	} {
		t.Run(display+" is a built-in role", func(t *testing.T) {
			info := memberInfoFrom(pulumiapi.Member{FGARole: &pulumiapi.FGARole{ID: "fga-id", Name: display}})
			require.NotNil(t, info.Role)
			assert.Equal(t, slug, *info.Role)
			assert.Nil(t, info.RoleId)
			require.NotNil(t, info.RoleName)
			assert.Equal(t, display, *info.RoleName)
		})
	}

	t.Run("a custom role is reported by ID", func(t *testing.T) {
		info := memberInfoFrom(pulumiapi.Member{FGARole: &pulumiapi.FGARole{ID: "fga-id", Name: "read-only-devops"}})
		assert.Nil(t, info.Role)
		require.NotNil(t, info.RoleId)
		assert.Equal(t, "fga-id", *info.RoleId)
	})
}
