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

// User-facing PermissionDescriptor maps use the wire-format `__type`
// discriminator at every level. Pulumi's Python SDK preserves
// `__`-prefixed keys across resource inputs as of pulumi/pulumi#22834
// (released in 3.235.0; pinned via the Python SDK's runtime
// requirement), so the language boundary is transparent and the
// provider passes the user's tree to the API verbatim.

package resources

import "fmt"

// validatePermissions sanity-checks a user-supplied descriptor map. The
// descriptor variants themselves are opaque to the provider (Allow,
// Group, Condition, Compose, IfThenElse, Select, the boolean operators,
// and any future variant Pulumi Cloud adds pass through unchanged); we
// only verify the top-level object carries the required `__type`
// discriminator key, so users see a clear error at preview rather than a
// 400 at apply.
func validatePermissions(node map[string]interface{}) error {
	if _, has := node["__type"]; !has {
		return fmt.Errorf("permissions descriptor missing required `__type` field")
	}
	return nil
}
