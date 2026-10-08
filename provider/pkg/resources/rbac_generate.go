// Copyright 2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

// zz_generated_rbac.go holds the RbacScope and RbacResourceType enums, generated
// from the Pulumi Cloud scope catalog. Refresh the catalog with
// scripts/fetch-rbac-scopes.sh, then run `go generate ./provider/pkg/resources/`.
//
//go:generate go run ../../tools/gen-rbac-scopes -scopes ../cloud/rbac-scopes.json -spec ../cloud/spec.json -out zz_generated_rbac.go
