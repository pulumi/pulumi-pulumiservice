// Copyright 2016-2020, Pulumi Corporation.
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

package main

import (
	_ "embed"

	psp "github.com/pulumi/pulumi-pulumiservice/provider/pkg/provider"
	"github.com/pulumi/pulumi/pkg/v3/resource/provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
	rpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

var providerName = "pulumiservice"

// embed schema.json directly into resource binary so that we can properly serve the schema
// directly from the resource provider
//
//go:embed schema.json
var schema string

// The version needs to be replaced using LDFLAGS on build
var Version string = "REPLACE_ON_BUILD"

func main() {
	// Start gRPC service for the pulumiservice provider
	err := provider.Main(providerName, func(host *provider.HostClient) (rpc.ResourceProviderServer, error) {
		return psp.MakeProvider(host, providerName, Version, schema)
	})
	if err != nil {
		cmdutil.ExitError(err.Error())
	}
}
