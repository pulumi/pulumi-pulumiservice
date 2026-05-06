// Copyright 2016-2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validatePermissions: top-level descriptor must carry `__type`. The
// rest of the descriptor is opaque to the provider — Pulumi Cloud is
// the source of truth for which variants exist.
func TestValidatePermissions(t *testing.T) {
	t.Parallel()
	t.Run("missing top-level __type", func(t *testing.T) {
		t.Parallel()
		err := validatePermissions(map[string]interface{}{
			"permissions": []interface{}{"stack:read"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "__type")
	})

	t.Run("present top-level __type", func(t *testing.T) {
		t.Parallel()
		err := validatePermissions(map[string]interface{}{
			"__type":      "PermissionDescriptorAllow",
			"permissions": []interface{}{"stack:read"},
		})
		require.NoError(t, err)
	})

	t.Run("accepts arbitrary __type values", func(t *testing.T) {
		t.Parallel()
		// validatePermissions does not gate descriptor variants — any
		// value passes the structural check, including future Cloud
		// additions the provider has no specific knowledge of.
		err := validatePermissions(map[string]interface{}{
			"__type": "PermissionDescriptorWhateverFutureCloudVariant",
		})
		require.NoError(t, err)
	})
}
