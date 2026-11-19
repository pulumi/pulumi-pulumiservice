// Copyright 2026, Pulumi Corporation.
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
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// autoNamedDescription is appended to the description of every auto-named field.
const autoNamedDescription = "If omitted, Pulumi generates a name from the resource's logical name plus a " +
	"random suffix, or follows the stack's " +
	"[autonaming configuration](https://www.pulumi.com/docs/iac/concepts/resources/names/#autonaming)."

// suffixedAutoName generates `<logical name>-<8 hex chars>`, shortening the
// logical name so the result is at most maxBytes bytes. Pulumi Cloud measures
// most name limits in bytes, and a name within maxBytes bytes is also within
// maxBytes characters. If clean is non-nil, it maps each rune of the logical
// name to one the API accepts.
func suffixedAutoName(maxBytes int, clean func(rune) rune) infer.AutoNameConfig {
	return infer.AutoNameConfig{
		Generate: func(_ context.Context, args infer.AutoNameGenerateArgs) (string, error) {
			suffix := fmt.Sprintf("-%08x", args.Rand.Uint32())
			base := args.URN.Name()
			if clean != nil {
				base = strings.Map(clean, base)
			}
			limit := maxBytes - len(suffix)
			for len(base) > limit {
				_, size := utf8.DecodeLastRuneInString(base)
				base = base[:len(base)-size]
			}
			return base + suffix, nil
		},
	}
}
