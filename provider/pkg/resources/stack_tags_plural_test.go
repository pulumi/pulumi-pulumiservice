package resources

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			if r.Method == http.MethodPost && r.URL.Path == "/api/stacks/org/project/stack/tags" {
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

		properties, err := util.ToProperties(input, "pulumi")
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

func TestStackTagsPluralRead(t *testing.T) {
	t.Run("Reads managed tags from stack", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && r.URL.Path == "/api/stacks/org/project/stack" {
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
		previousProps, err := util.ToProperties(previousInput, "pulumi")
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
		err = util.FromPropertyMap(props, "pulumi", &output)
		require.NoError(t, err)

		assert.Equal(t, 2, len(output.Tags))
		assert.Equal(t, "value1", output.Tags["tag1"])
		assert.Equal(t, "value2", output.Tags["tag2"])
		assert.NotContains(t, output.Tags, "other-tag")
	})
}

func TestStackTagsPluralUpdate(t *testing.T) {
	t.Run("Adds, removes, and modifies tags", func(t *testing.T) {
		var createdTags []pulumiapi.StackTag
		var deletedTags []string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost && r.URL.Path == "/api/stacks/org/project/stack/tags" {
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

		oldProps, err := util.ToProperties(oldInput, "pulumi")
		require.NoError(t, err)
		newProps, err := util.ToProperties(newInput, "pulumi")
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

		properties, err := util.ToProperties(input, "pulumi")
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

		props, err := util.ToProperties(input, "pulumi")
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

		oldProps, err := util.ToProperties(oldInput, "pulumi")
		require.NoError(t, err)
		newProps, err := util.ToProperties(newInput, "pulumi")
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

		oldProps, err := util.ToProperties(oldInput, "pulumi")
		require.NoError(t, err)
		newProps, err := util.ToProperties(newInput, "pulumi")
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

		props, err := util.ToProperties(input, "pulumi")
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

		// Convert back from PropertyMap
		st := &PulumiServiceStackTagsResource{}
		output := st.ToPulumiServiceStackTagsInput(propMap)
		assert.Equal(t, input.Organization, output.Organization)
		assert.Equal(t, input.Project, output.Project)
		assert.Equal(t, input.Stack, output.Stack)
		assert.Equal(t, input.Tags, output.Tags)
	})
}
