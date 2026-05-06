// Copyright 2016-2026, Pulumi Corporation.
package cloud_test

import (
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/cloud"
)

// TestPreviewEnvironmentsTokensUseEscPreviewModule pins the user-facing token
// shape for the four preview/environments-derived resources. The metadata
// keys keep their _preview_environments suffix (canonical OpenAPI-derived
// form), but the user-facing `token` field maps them into the escPreview
// module — see scaffold-metadata/main.go moduleAliases.
func TestPreviewEnvironmentsTokensUseEscPreviewModule(t *testing.T) {
	md := cloud.Metadata()

	cases := map[string]string{
		"pulumiservice:v2:Environment_preview_environments":    "pulumiservice:v2/escPreview:Environment",
		"pulumiservice:v2:EnvironmentTag_preview_environments": "pulumiservice:v2/escPreview:EnvironmentTag",
		"pulumiservice:v2:RevisionTag_preview_environments":    "pulumiservice:v2/escPreview:RevisionTag",
		"pulumiservice:v2:Webhook_preview_environments":        "pulumiservice:v2/escPreview:Webhook",
	}
	for key, want := range cases {
		rm, ok := md.Resources[key]
		if !ok {
			t.Errorf("missing metadata entry %q", key)
			continue
		}
		if rm.Token != want {
			t.Errorf("metadata[%q].token = %q, want %q", key, rm.Token, want)
		}
	}
}

// TestPreviewEnvironmentsAliasOldToken ensures stack state under the old
// pulumiservice:v2/preview:* tokens rebinds via Pulumi schema aliases.
func TestPreviewEnvironmentsAliasOldToken(t *testing.T) {
	md := cloud.Metadata()
	cases := map[string]string{
		"pulumiservice:v2:Environment_preview_environments":    "pulumiservice:v2/preview:Environment",
		"pulumiservice:v2:EnvironmentTag_preview_environments": "pulumiservice:v2/preview:EnvironmentTag",
		"pulumiservice:v2:RevisionTag_preview_environments":    "pulumiservice:v2/preview:RevisionTag",
		"pulumiservice:v2:Webhook_preview_environments":        "pulumiservice:v2/preview:Webhook",
	}
	for key, want := range cases {
		rm, ok := md.Resources[key]
		if !ok {
			t.Errorf("missing metadata entry %q", key)
			continue
		}
		var found bool
		for _, a := range rm.Aliases {
			if a == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("metadata[%q].aliases missing %q (got %v)", key, want, rm.Aliases)
		}
	}
}
