package resources

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/rpcutil/rpcerror"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

const (
	testStackPath     = "/api/stacks/org/project/stack"
	testStackTagsPath = "/api/stacks/org/project/stack/tags"
)

// extractTagNameFromPath extracts the tag name from a URL path like /api/stacks/org/project/stack/tags/tagname
func extractTagNameFromPath(urlPath string) string {
	return path.Base(urlPath)
}

func TestStackTagsPluralCreate(t *testing.T) {
	t.Run("Creates all tags in the map", func(t *testing.T) {
		var createdTags []pulumiapi.StackTag
		var requestCount int

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			if r.Method == http.MethodPost && r.URL.Path == testStackTagsPath {
				var tag pulumiapi.StackTag
				err := json.NewDecoder(r.Body).Decode(&tag)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				createdTags = append(createdTags, tag)
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		apiClient, err := pulumiapi.NewClient(server.Client(), "", server.URL)
		require.NoError(t, err)

		st := &PulumiServiceStackTagsResource{
			Client: apiClient,
		}

		input := PulumiServiceStackTagsInput{
			Organization: "org",
			Project:      "project",
			Stack:        "stack",
			Tags: map[string]string{
				"tag1": "value1",
				"tag2": "value2",
				"tag3": "value3",
			},
		}

		properties, err := util.ToProperties(input)
		require.NoError(t, err)

		createReq := pulumirpc.CreateRequest{
			Properties: properties,
		}

		resp, err := st.Create(&createReq)
		require.NoError(t, err)
		assert.Equal(t, "org/project/stack/tags", resp.Id)
		assert.Equal(t, 3, len(createdTags))

		// Verify all tags were created
		tagMap := make(map[string]string)
		for _, tag := range createdTags {
			tagMap[tag.Name] = tag.Value
		}
		assert.Equal(t, "value1", tagMap["tag1"])
		assert.Equal(t, "value2", tagMap["tag2"])
		assert.Equal(t, "value3", tagMap["tag3"])
	})
}

func TestStackTagsPluralCreatePartialFailure(t *testing.T) {
	t.Run("Returns partial state when a tag creation fails mid-way", func(t *testing.T) {
		// Fail the third tag (tags are created in sorted order: tag1, tag2, tag3).
		var created []string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			var tag pulumiapi.StackTag
			if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if tag.Name == "tag3" {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			created = append(created, tag.Name)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		apiClient, err := pulumiapi.NewClient(server.Client(), "", server.URL)
		require.NoError(t, err)

		st := &PulumiServiceStackTagsResource{Client: apiClient}

		input := PulumiServiceStackTagsInput{
			Organization: "org",
			Project:      "project",
			Stack:        "stack",
			Tags: map[string]string{
				"tag1": "value1",
				"tag2": "value2",
				"tag3": "value3",
			},
		}
		properties, err := util.ToProperties(input)
		require.NoError(t, err)

		_, err = st.Create(&pulumirpc.CreateRequest{Properties: properties})
		require.Error(t, err)
		assert.ElementsMatch(t, []string{"tag1", "tag2"}, created)

		// The error must carry ErrorResourceInitFailed with the successfully-created
		// subset so Pulumi records them in state and doesn't recreate them on retry.
		rpcErr, ok := rpcerror.FromError(err)
		require.True(t, ok, "expected a pulumi rpcerror")

		var initFailed *pulumirpc.ErrorResourceInitFailed
		for _, d := range rpcErr.Details() {
			if f, ok := d.(*pulumirpc.ErrorResourceInitFailed); ok {
				initFailed = f
				break
			}
		}
		require.NotNil(t, initFailed, "expected ErrorResourceInitFailed detail")
		assert.Equal(t, "org/project/stack/tags", initFailed.Id)

		partialProps, err := plugin.UnmarshalProperties(
			initFailed.Properties, plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
		require.NoError(t, err)
		var partial PulumiServiceStackTagsInput
		require.NoError(t, util.FromPropertyMap(partialProps, &partial))
		assert.Equal(t, map[string]string{"tag1": "value1", "tag2": "value2"}, partial.Tags)
	})
}

func TestStackTagsPluralRead(t *testing.T) {
	t.Run("Reads managed tags from stack", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && r.URL.Path == testStackPath {
				response := map[string]interface{}{
					"tags": map[string]string{
						"tag1":      "value1",
						"tag2":      "value2",
						"other-tag": "other-value",
					},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(response)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		apiClient, err := pulumiapi.NewClient(server.Client(), "", server.URL)
		require.NoError(t, err)

		st := &PulumiServiceStackTagsResource{
			Client: apiClient,
		}

		// Simulate previous state with tag1 and tag2 managed
		previousInput := PulumiServiceStackTagsInput{
			Organization: "org",
			Project:      "project",
			Stack:        "stack",
			Tags: map[string]string{
				"tag1": "value1",
				"tag2": "value2",
			},
		}
		previousProps, err := util.ToProperties(previousInput)
		require.NoError(t, err)

		readReq := pulumirpc.ReadRequest{
			Id:     "org/project/stack/tags",
			Inputs: previousProps,
		}

		resp, err := st.Read(&readReq)
		require.NoError(t, err)
		assert.Equal(t, "org/project/stack/tags", resp.Id)

		// Verify only managed tags are returned (not other-tag)
		var output PulumiServiceStackTagsInput
		props, err := plugin.UnmarshalProperties(resp.Properties, plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
		require.NoError(t, err)
		err = util.FromPropertyMap(props, &output)
		require.NoError(t, err)

		assert.Equal(t, 2, len(output.Tags))
		assert.Equal(t, "value1", output.Tags["tag1"])
		assert.Equal(t, "value2", output.Tags["tag2"])
		assert.NotContains(t, output.Tags, "other-tag")
	})

	t.Run("Import adopts every tag on the stack when there is no prior state", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && r.URL.Path == testStackPath {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"tags": map[string]string{
						"env":    "prod",
						"team":   "platform",
						"region": "us-east-1",
					},
				})
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		apiClient, err := pulumiapi.NewClient(server.Client(), "", server.URL)
		require.NoError(t, err)
		st := &PulumiServiceStackTagsResource{Client: apiClient}

		// req.Inputs is nil — the pulumi import path.
		resp, err := st.Read(&pulumirpc.ReadRequest{Id: "org/project/stack/tags"})
		require.NoError(t, err)
		assert.Equal(t, "org/project/stack/tags", resp.Id)

		props, err := plugin.UnmarshalProperties(
			resp.Properties, plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
		require.NoError(t, err)
		var output PulumiServiceStackTagsInput
		require.NoError(t, util.FromPropertyMap(props, &output))

		assert.Equal(t, map[string]string{
			"env":    "prod",
			"team":   "platform",
			"region": "us-east-1",
		}, output.Tags)
	})

	t.Run("Returns an error when stored inputs can't be decoded", func(t *testing.T) {
		// No HTTP call should reach the server — Read must fail before talking to the API.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Still satisfy the GET since Read calls GetStackTags before decoding inputs.
			if r.Method == http.MethodGet && r.URL.Path == testStackPath {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"tags": map[string]string{}})
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		apiClient, err := pulumiapi.NewClient(server.Client(), "", server.URL)
		require.NoError(t, err)
		st := &PulumiServiceStackTagsResource{Client: apiClient}

		// Hand-craft inputs whose `tags` field is a string rather than a map.
		// FromPropertyMap should reject the type mismatch.
		malformed := resource.PropertyMap{
			"organization": resource.NewStringProperty("org"),
			"project":      resource.NewStringProperty("project"),
			"stack":        resource.NewStringProperty("stack"),
			"tags":         resource.NewStringProperty("not-a-map"),
		}
		props, err := plugin.MarshalProperties(malformed, plugin.MarshalOptions{})
		require.NoError(t, err)

		_, err = st.Read(&pulumirpc.ReadRequest{Id: "org/project/stack/tags", Inputs: props})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode stored inputs")
	})
}

func TestStackTagsPluralUpdate(t *testing.T) {
	t.Run("Adds, removes, and modifies tags", func(t *testing.T) {
		var createdTags []pulumiapi.StackTag
		var deletedTags []string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost && r.URL.Path == testStackTagsPath {
				var tag pulumiapi.StackTag
				err := json.NewDecoder(r.Body).Decode(&tag)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				createdTags = append(createdTags, tag)
				w.WriteHeader(http.StatusOK)
			} else if r.Method == http.MethodDelete {
				tagName := extractTagNameFromPath(r.URL.Path)
				if tagName != "" {
					deletedTags = append(deletedTags, tagName)
				}
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		apiClient, err := pulumiapi.NewClient(server.Client(), "", server.URL)
		require.NoError(t, err)

		st := &PulumiServiceStackTagsResource{
			Client: apiClient,
		}

		oldInput := PulumiServiceStackTagsInput{
			Organization: "org",
			Project:      "project",
			Stack:        "stack",
			Tags: map[string]string{
				"tag1": "value1",
				"tag2": "value2",
				"tag3": "value3",
			},
		}

		newInput := PulumiServiceStackTagsInput{
			Organization: "org",
			Project:      "project",
			Stack:        "stack",
			Tags: map[string]string{
				"tag1": "new-value1", // Modified
				"tag2": "value2",     // Unchanged
				"tag4": "value4",     // Added
				// tag3 removed
			},
		}

		oldProps, err := util.ToProperties(oldInput)
		require.NoError(t, err)
		newProps, err := util.ToProperties(newInput)
		require.NoError(t, err)

		updateReq := pulumirpc.UpdateRequest{
			Olds: oldProps,
			News: newProps,
		}

		resp, err := st.Update(&updateReq)
		require.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify tag3 was deleted
		assert.Contains(t, deletedTags, "tag3")

		// Verify tag1 was deleted (for update) and recreated with new value
		assert.Contains(t, deletedTags, "tag1")

		// Verify new tags were created
		createdMap := make(map[string]string)
		for _, tag := range createdTags {
			createdMap[tag.Name] = tag.Value
		}
		assert.Equal(t, "new-value1", createdMap["tag1"])
		assert.Equal(t, "value4", createdMap["tag4"])
	})
}

func TestStackTagsPluralUpdatePartialFailure(t *testing.T) {
	// Helper: build a server that fails when a specific tag is created or deleted.
	// Tags are processed in sorted order by the resource, so the failure point is
	// predictable.
	newServer := func(failDelete, failCreate string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && r.URL.Path == testStackTagsPath:
				var tag pulumiapi.StackTag
				if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if tag.Name == failCreate {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			case r.Method == http.MethodDelete:
				if extractTagNameFromPath(r.URL.Path) == failDelete {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
	}

	extractPartialState := func(t *testing.T, err error) PulumiServiceStackTagsInput {
		t.Helper()
		rpcErr, ok := rpcerror.FromError(err)
		require.True(t, ok, "expected a pulumi rpcerror")
		var initFailed *pulumirpc.ErrorResourceInitFailed
		for _, d := range rpcErr.Details() {
			if f, ok := d.(*pulumirpc.ErrorResourceInitFailed); ok {
				initFailed = f
				break
			}
		}
		require.NotNil(t, initFailed, "expected ErrorResourceInitFailed detail")
		partialProps, err := plugin.UnmarshalProperties(
			initFailed.Properties, plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
		require.NoError(t, err)
		var partial PulumiServiceStackTagsInput
		require.NoError(t, util.FromPropertyMap(partialProps, &partial))
		return partial
	}

	oldInput := PulumiServiceStackTagsInput{
		Organization: "org",
		Project:      "project",
		Stack:        "stack",
		Tags: map[string]string{
			"tag1": "v1",
			"tag2": "v2",
			"tag3": "v3",
		},
	}
	// Remove tag1, change tag2, add tag4. Sorted deletes: [tag1, tag2]; sorted
	// creates: [tag2, tag4].
	newInput := PulumiServiceStackTagsInput{
		Organization: "org",
		Project:      "project",
		Stack:        "stack",
		Tags: map[string]string{
			"tag2": "v2-new",
			"tag3": "v3",
			"tag4": "v4",
		},
	}

	t.Run("fails mid-delete: state reflects tags still on the server", func(t *testing.T) {
		// Fail on delete of tag2 (second in sorted order). tag1 was already deleted.
		server := newServer("tag2", "")
		defer server.Close()

		apiClient, err := pulumiapi.NewClient(server.Client(), "", server.URL)
		require.NoError(t, err)
		st := &PulumiServiceStackTagsResource{Client: apiClient}

		oldProps, err := util.ToProperties(oldInput)
		require.NoError(t, err)
		newProps, err := util.ToProperties(newInput)
		require.NoError(t, err)

		_, err = st.Update(&pulumirpc.UpdateRequest{Olds: oldProps, News: newProps})
		require.Error(t, err)

		partial := extractPartialState(t, err)
		// tag1 was deleted; tag2 delete failed, so tag2 is still live; tag3 was
		// untouched. No creates happened yet.
		assert.Equal(t, map[string]string{"tag2": "v2", "tag3": "v3"}, partial.Tags)
	})

	t.Run("fails mid-create: state reflects deletes plus successful creates", func(t *testing.T) {
		// Fail on create of tag4 (second in sorted creates). By that point:
		// deletes tag1, tag2 succeeded; create tag2="v2-new" succeeded.
		server := newServer("", "tag4")
		defer server.Close()

		apiClient, err := pulumiapi.NewClient(server.Client(), "", server.URL)
		require.NoError(t, err)
		st := &PulumiServiceStackTagsResource{Client: apiClient}

		oldProps, err := util.ToProperties(oldInput)
		require.NoError(t, err)
		newProps, err := util.ToProperties(newInput)
		require.NoError(t, err)

		_, err = st.Update(&pulumirpc.UpdateRequest{Olds: oldProps, News: newProps})
		require.Error(t, err)

		partial := extractPartialState(t, err)
		assert.Equal(t, map[string]string{"tag2": "v2-new", "tag3": "v3"}, partial.Tags)
	})
}

func TestStackTagsPluralDelete(t *testing.T) {
	t.Run("Deletes all managed tags", func(t *testing.T) {
		var deletedTags []string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodDelete {
				tagName := extractTagNameFromPath(r.URL.Path)
				if tagName != "" {
					deletedTags = append(deletedTags, tagName)
				}
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		apiClient, err := pulumiapi.NewClient(server.Client(), "", server.URL)
		require.NoError(t, err)

		st := &PulumiServiceStackTagsResource{
			Client: apiClient,
		}

		input := PulumiServiceStackTagsInput{
			Organization: "org",
			Project:      "project",
			Stack:        "stack",
			Tags: map[string]string{
				"tag1": "value1",
				"tag2": "value2",
				"tag3": "value3",
			},
		}

		properties, err := util.ToProperties(input)
		require.NoError(t, err)

		deleteReq := pulumirpc.DeleteRequest{
			Properties: properties,
		}

		_, err = st.Delete(&deleteReq)
		require.NoError(t, err)

		// Verify all tags were deleted
		assert.Equal(t, 3, len(deletedTags))
		assert.Contains(t, deletedTags, "tag1")
		assert.Contains(t, deletedTags, "tag2")
		assert.Contains(t, deletedTags, "tag3")
	})
}

func TestStackTagsPluralDiff(t *testing.T) {
	t.Run("Detects no changes when tags are identical", func(t *testing.T) {
		st := &PulumiServiceStackTagsResource{}

		input := PulumiServiceStackTagsInput{
			Organization: "org",
			Project:      "project",
			Stack:        "stack",
			Tags: map[string]string{
				"tag1": "value1",
				"tag2": "value2",
			},
		}

		props, err := util.ToProperties(input)
		require.NoError(t, err)

		diffReq := pulumirpc.DiffRequest{
			OldInputs: props,
			News:      props,
		}

		resp, err := st.Diff(&diffReq)
		require.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes)
	})

	t.Run("Detects changes when tags are modified", func(t *testing.T) {
		st := &PulumiServiceStackTagsResource{}

		oldInput := PulumiServiceStackTagsInput{
			Organization: "org",
			Project:      "project",
			Stack:        "stack",
			Tags: map[string]string{
				"tag1": "value1",
			},
		}

		newInput := PulumiServiceStackTagsInput{
			Organization: "org",
			Project:      "project",
			Stack:        "stack",
			Tags: map[string]string{
				"tag1": "new-value1",
			},
		}

		oldProps, err := util.ToProperties(oldInput)
		require.NoError(t, err)
		newProps, err := util.ToProperties(newInput)
		require.NoError(t, err)

		diffReq := pulumirpc.DiffRequest{
			OldInputs: oldProps,
			News:      newProps,
		}

		resp, err := st.Diff(&diffReq)
		require.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		assert.NotNil(t, resp.DetailedDiff)
	})

	t.Run("Requires replacement when organization changes", func(t *testing.T) {
		st := &PulumiServiceStackTagsResource{}

		oldInput := PulumiServiceStackTagsInput{
			Organization: "org1",
			Project:      "project",
			Stack:        "stack",
			Tags: map[string]string{
				"tag1": "value1",
			},
		}

		newInput := PulumiServiceStackTagsInput{
			Organization: "org2",
			Project:      "project",
			Stack:        "stack",
			Tags: map[string]string{
				"tag1": "value1",
			},
		}

		oldProps, err := util.ToProperties(oldInput)
		require.NoError(t, err)
		newProps, err := util.ToProperties(newInput)
		require.NoError(t, err)

		diffReq := pulumirpc.DiffRequest{
			OldInputs: oldProps,
			News:      newProps,
		}

		resp, err := st.Diff(&diffReq)
		require.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		assert.True(t, resp.DeleteBeforeReplace)
		assert.Contains(t, resp.Replaces, "organization")
	})
}

func TestStackTagsPluralCheck(t *testing.T) {
	t.Run("Returns inputs without failures", func(t *testing.T) {
		st := &PulumiServiceStackTagsResource{}

		input := PulumiServiceStackTagsInput{
			Organization: "org",
			Project:      "project",
			Stack:        "stack",
			Tags: map[string]string{
				"tag1": "value1",
			},
		}

		props, err := util.ToProperties(input)
		require.NoError(t, err)

		checkReq := pulumirpc.CheckRequest{
			News: props,
		}

		resp, err := st.Check(&checkReq)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Nil(t, resp.Failures)
		assert.Equal(t, props, resp.Inputs)
	})
}

func TestStackTagsPluralName(t *testing.T) {
	t.Run("Returns correct resource name", func(t *testing.T) {
		st := &PulumiServiceStackTagsResource{}
		assert.Equal(t, "pulumiservice:index:StackTags", st.Name())
	})
}

func TestStackTagsPluralPropertyMapConversion(t *testing.T) {
	t.Run("Converts to and from PropertyMap correctly", func(t *testing.T) {
		input := PulumiServiceStackTagsInput{
			Organization: "org",
			Project:      "project",
			Stack:        "stack",
			Tags: map[string]string{
				"tag1": "value1",
				"tag2": "value2",
			},
		}

		// Convert to PropertyMap
		propMap := input.ToPropertyMap()
		assert.NotNil(t, propMap)
		assert.True(t, propMap["organization"].IsString())
		assert.Equal(t, "org", propMap["organization"].StringValue())

		// Convert back from PropertyMap via the shared util helper.
		var output PulumiServiceStackTagsInput
		require.NoError(t, util.FromPropertyMap(propMap, &output))
		assert.Equal(t, input.Organization, output.Organization)
		assert.Equal(t, input.Project, output.Project)
		assert.Equal(t, input.Stack, output.Stack)
		assert.Equal(t, input.Tags, output.Tags)
	})
}
