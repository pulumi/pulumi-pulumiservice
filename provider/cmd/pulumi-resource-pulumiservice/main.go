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
	"os"
	"os/signal"
	"runtime/coverage"
	"syscall"

	"github.com/pulumi/pulumi/pkg/v3/resource/provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
	rpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	psp "github.com/pulumi/pulumi-pulumiservice/provider/pkg/provider"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/version"
)

var providerName = "pulumiservice"

// The version needs to be replaced using LDFLAGS on build
var Version = version.Version

func main() {
	// When the binary is built with `go build -cover`, the Go runtime only writes
	// covcounters.* files at clean process exit. The Pulumi engine signals
	// providers with SIGINT/SIGTERM at deployment teardown, so without this
	// hook the integration test workflow gets covmeta files but no counters
	// and `go tool covdata textfmt` produces an empty profile. Calling
	// flushCoverage on signal lets the counters land before we exit.
	flushCoverage := installCoverFlushOnSignal()

	// Start gRPC service for the pulumiservice provider
	err := provider.Main(providerName, func(host *provider.HostClient) (rpc.ResourceProviderServer, error) {
		return psp.MakeProvider(host, providerName, Version)
	})
	flushCoverage()
	if err != nil {
		cmdutil.ExitError(err.Error())
	}
}

// installCoverFlushOnSignal returns a function that flushes coverage counters
// (if `GOCOVERDIR` is set) and is also registered to run on SIGINT/SIGTERM.
// The returned function is idempotent and safe to call from `main` after
// `provider.Main` returns.
func installCoverFlushOnSignal() func() {
	flushed := make(chan struct{})
	flush := func() {
		select {
		case <-flushed:
			return
		default:
			close(flushed)
		}
		writeCoverage()
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		flush()
		os.Exit(0)
	}()

	return flush
}

// writeCoverage flushes coverage counters to GOCOVERDIR. It is a no-op if
// GOCOVERDIR is unset or the binary was not built with `-cover`.
func writeCoverage() {
	dir := os.Getenv("GOCOVERDIR")
	if dir == "" {
		return
	}
	// Errors here only mean the binary wasn't built with -cover or the dir
	// doesn't exist; either way there's nothing useful to log to stderr.
	_ = coverage.WriteCountersDir(dir)
}
