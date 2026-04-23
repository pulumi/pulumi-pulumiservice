// Copyright 2016-2026, Pulumi Corporation.
//
// pulumi-resource-pulumiservice is the Pulumi Service Provider v2 binary.
// It embeds the generated schema.json and metadata.json emitted by
// pulumi-gen-pulumiservice and serves Pulumi's gRPC protocol by delegating
// every CRUD call to runtime.Dispatcher. v2 has zero hand-coded resources;
// every supported resource is expressed in provider/resource-map.yaml.

package main

import (
	_ "embed"

	"github.com/pulumi/pulumi/pkg/v3/resource/provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
	rpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	pspprov "github.com/pulumi/pulumi-pulumiservice/provider/pkg/provider"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/version"
)

// Embedded at build time from the generator's bin/ output. `make v2_gen`
// must have run first.
//
//go:embed schema.json
var schemaBytes []byte

//go:embed metadata.json
var metadataBytes []byte

var providerName = "pulumiservice"

// Overridden via LDFLAGS at build time.
var Version = version.Version

func main() {
	err := provider.Main(providerName, func(host *provider.HostClient) (rpc.ResourceProviderServer, error) {
		return pspprov.New(providerName, Version, schemaBytes, metadataBytes)
	})
	if err != nil {
		cmdutil.ExitError(err.Error())
	}
}
