// Copyright 2016-2026, Pulumi Corporation.
//
// pulumi-resource-pulumiservice is the Pulumi Service Provider v2 binary.
// It is built directly from the OpenAPI spec and resource map embedded
// at provider/pkg/embedded — `go build ./provider/cmd/pulumi-resource-pulumiservice`
// produces a runnable provider with no separate code-generation step.
//
// The runtime is built on github.com/pulumi/pulumi-go-provider; CRUD
// dispatch is metadata-driven (see provider/pkg/runtime). v2 has zero
// hand-coded resources; every supported resource is expressed in
// resource-map.yaml.

package main

import (
	"context"

	pgo "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/embedded"
	pspprov "github.com/pulumi/pulumi-pulumiservice/provider/pkg/provider"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/version"
)

const providerName = "pulumiservice"

// Overridden via LDFLAGS at build time. Package-local var rather than
// re-exported so we keep the old build pipeline (LDFLAGS_PROJ_VERSION
// targets provider/pkg/version.Version) wired through unchanged.
var Version = version.Version

func main() {
	prov, err := pspprov.New(embedded.Spec(), embedded.ResourceMap())
	if err != nil {
		cmdutil.ExitError(err.Error())
	}
	if err := pgo.RunProvider(context.Background(), providerName, Version, prov); err != nil {
		cmdutil.ExitError(err.Error())
	}
}
