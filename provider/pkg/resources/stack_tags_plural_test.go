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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

// stackTagsClientMock implements the slice of config.Client that StackTags uses.
type stackTagsClientMock struct {
	config.Client
	createStackTagFunc func(ctx context.Context, stack pulumiapi.StackIdentifier, tag pulumiapi.StackTag) error
	deleteStackTagFunc func(ctx context.Context, stack pulumiapi.StackIdentifier, tagName string) error
	getStackTagsFunc   func(ctx context.Context, stack pulumiapi.StackIdentifier) (map[string]string, error)
}

func (m *stackTagsClientMock) CreateStackTag(
	ctx context.Context, stack pulumiapi.StackIdentifier, tag pulumiapi.StackTag,
) error {
	return m.createStackTagFunc(ctx, stack, tag)
}

func (m *stackTagsClientMock) DeleteStackTag(
	ctx context.Context, stack pulumiapi.StackIdentifier, tagName string,
) error {
	return m.deleteStackTagFunc(ctx, stack, tagName)
}

func (m *stackTagsClientMock) GetStackTags(
	ctx context.Context, stack pulumiapi.StackIdentifier,
) (map[string]string, error) {
	return m.getStackTagsFunc(ctx, stack)
}

func TestStackTagsResourceID(t *testing.T) {
	t.Run("formats id", func(t *testing.T) {
		assert.Equal(t, "org/proj/stk/tags", stackTagsResourceID("org", "proj", "stk"))
	})
}

func TestParseStackTagsResourceID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		org, proj, stk, err := parseStackTagsResourceID("org/proj/stk/tags")
		require.NoError(t, err)
		assert.Equal(t, "org", org)
		assert.Equal(t, "proj", proj)
		assert.Equal(t, "stk", stk)
	})

	t.Run("missing tags suffix", func(t *testing.T) {
		_, _, _, err := parseStackTagsResourceID("org/proj/stk")
		require.Error(t, err)
	})

	t.Run("wrong suffix", func(t *testing.T) {
		_, _, _, err := parseStackTagsResourceID("org/proj/stk/notags")
		require.Error(t, err)
	})

	t.Run("too many parts", func(t *testing.T) {
		_, _, _, err := parseStackTagsResourceID("org/proj/stk/tags/extra")
		require.Error(t, err)
	})
}

func TestStackTagsCreate(t *testing.T) {
	t.Run("creates all tags in sorted order", func(t *testing.T) {
		var created []pulumiapi.StackTag
		mock := &stackTagsClientMock{
			createStackTagFunc: func(_ context.Context, _ pulumiapi.StackIdentifier, tag pulumiapi.StackTag) error {
				created = append(created, tag)
				return nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)

		resp, err := (&StackTags{}).Create(ctx, infer.CreateRequest[StackTagsInput]{
			Inputs: StackTagsInput{
				Organization: "org",
				Project:      "project",
				Stack:        "stack",
				Tags:         map[string]string{"b": "2", "a": "1", "c": "3"},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "org/project/stack/tags", resp.ID)
		assert.Equal(t, []pulumiapi.StackTag{
			{Name: "a", Value: "1"},
			{Name: "b", Value: "2"},
			{Name: "c", Value: "3"},
		}, created)
	})

	t.Run("dry-run returns inputs and skips API calls", func(t *testing.T) {
		mock := &stackTagsClientMock{
			createStackTagFunc: func(context.Context, pulumiapi.StackIdentifier, pulumiapi.StackTag) error {
				t.Fatal("Create should not call API during dry run")
				return nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)

		resp, err := (&StackTags{}).Create(ctx, infer.CreateRequest[StackTagsInput]{
			DryRun: true,
			Inputs: StackTagsInput{
				Organization: "org",
				Project:      "project",
				Stack:        "stack",
				Tags:         map[string]string{"a": "1"},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "org/project/stack/tags", resp.ID)
		assert.Equal(t, StackTagsState{
			Organization: "org",
			Project:      "project",
			Stack:        "stack",
			Tags:         map[string]string{"a": "1"},
		}, resp.Output)
	})

	t.Run("partial failure returns ResourceInitFailedError with successful tags", func(t *testing.T) {
		mock := &stackTagsClientMock{
			createStackTagFunc: func(_ context.Context, _ pulumiapi.StackIdentifier, tag pulumiapi.StackTag) error {
				if tag.Name == "tag3" {
					return errors.New("boom")
				}
				return nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)

		resp, err := (&StackTags{}).Create(ctx, infer.CreateRequest[StackTagsInput]{
			Inputs: StackTagsInput{
				Organization: "org",
				Project:      "project",
				Stack:        "stack",
				Tags:         map[string]string{"tag1": "v1", "tag2": "v2", "tag3": "v3"},
			},
		})
		require.Error(t, err)

		var initFailed infer.ResourceInitFailedError
		require.True(t, errors.As(err, &initFailed))
		require.Len(t, initFailed.Reasons, 1)
		assert.Contains(t, initFailed.Reasons[0], "tag3")
		assert.Equal(t, "org/project/stack/tags", resp.ID)
		assert.Equal(t, StackTagsState{
			Organization: "org",
			Project:      "project",
			Stack:        "stack",
			Tags:         map[string]string{"tag1": "v1", "tag2": "v2"},
		}, resp.Output)
	})
}

func TestStackTagsUpdate(t *testing.T) {
	oldState := StackTagsState{
		Organization: "org",
		Project:      "project",
		Stack:        "stack",
		Tags:         map[string]string{"tag1": "v1", "tag2": "v2", "tag3": "v3"},
	}
	// Remove tag1, change tag2, add tag4.
	newInputs := StackTagsInput{
		Organization: "org",
		Project:      "project",
		Stack:        "stack",
		Tags:         map[string]string{"tag2": "v2-new", "tag3": "v3", "tag4": "v4"},
	}

	t.Run("applies add/delete/modify operations", func(t *testing.T) {
		var deletes []string
		var creates []pulumiapi.StackTag
		mock := &stackTagsClientMock{
			createStackTagFunc: func(_ context.Context, _ pulumiapi.StackIdentifier, tag pulumiapi.StackTag) error {
				creates = append(creates, tag)
				return nil
			},
			deleteStackTagFunc: func(_ context.Context, _ pulumiapi.StackIdentifier, tagName string) error {
				deletes = append(deletes, tagName)
				return nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)

		resp, err := (&StackTags{}).Update(ctx, infer.UpdateRequest[StackTagsInput, StackTagsState]{
			State:  oldState,
			Inputs: newInputs,
		})
		require.NoError(t, err)
		// Deletes are sorted: tag1 (removed), tag2 (modified).
		assert.Equal(t, []string{"tag1", "tag2"}, deletes)
		// Creates are sorted: tag2 (modified), tag4 (added).
		assert.Equal(t, []pulumiapi.StackTag{
			{Name: "tag2", Value: "v2-new"},
			{Name: "tag4", Value: "v4"},
		}, creates)
		assert.Equal(t, newInputs, resp.Output)
	})

	t.Run("dry-run skips API and returns inputs", func(t *testing.T) {
		mock := &stackTagsClientMock{
			createStackTagFunc: func(context.Context, pulumiapi.StackIdentifier, pulumiapi.StackTag) error {
				t.Fatal("Update should not call API during dry run")
				return nil
			},
			deleteStackTagFunc: func(context.Context, pulumiapi.StackIdentifier, string) error {
				t.Fatal("Update should not call API during dry run")
				return nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)

		resp, err := (&StackTags{}).Update(ctx, infer.UpdateRequest[StackTagsInput, StackTagsState]{
			DryRun: true,
			State:  oldState,
			Inputs: newInputs,
		})
		require.NoError(t, err)
		assert.Equal(t, newInputs, resp.Output)
	})

	t.Run("delete failure mid-update returns partial state with surviving tags", func(t *testing.T) {
		mock := &stackTagsClientMock{
			deleteStackTagFunc: func(_ context.Context, _ pulumiapi.StackIdentifier, tagName string) error {
				if tagName == "tag2" {
					return errors.New("delete boom")
				}
				return nil
			},
			createStackTagFunc: func(context.Context, pulumiapi.StackIdentifier, pulumiapi.StackTag) error {
				t.Fatal("creates should not run after a delete failure")
				return nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)

		resp, err := (&StackTags{}).Update(ctx, infer.UpdateRequest[StackTagsInput, StackTagsState]{
			State:  oldState,
			Inputs: newInputs,
		})
		require.Error(t, err)
		var initFailed infer.ResourceInitFailedError
		require.True(t, errors.As(err, &initFailed))
		// tag1 was successfully deleted; tag2 failed and is still live; tag3
		// has not been touched yet.
		assert.Equal(t, StackTagsState{
			Organization: "org",
			Project:      "project",
			Stack:        "stack",
			Tags:         map[string]string{"tag2": "v2", "tag3": "v3"},
		}, resp.Output)
	})

	t.Run("create failure mid-update returns partial state with deletes plus successful creates", func(t *testing.T) {
		mock := &stackTagsClientMock{
			deleteStackTagFunc: func(context.Context, pulumiapi.StackIdentifier, string) error {
				return nil
			},
			createStackTagFunc: func(_ context.Context, _ pulumiapi.StackIdentifier, tag pulumiapi.StackTag) error {
				if tag.Name == "tag4" {
					return errors.New("create boom")
				}
				return nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)

		resp, err := (&StackTags{}).Update(ctx, infer.UpdateRequest[StackTagsInput, StackTagsState]{
			State:  oldState,
			Inputs: newInputs,
		})
		require.Error(t, err)
		var initFailed infer.ResourceInitFailedError
		require.True(t, errors.As(err, &initFailed))
		// tag1, tag2 deletes succeeded; tag2 was recreated with the new value;
		// tag3 is untouched; tag4 create failed.
		assert.Equal(t, StackTagsState{
			Organization: "org",
			Project:      "project",
			Stack:        "stack",
			Tags:         map[string]string{"tag2": "v2-new", "tag3": "v3"},
		}, resp.Output)
	})
}

func TestStackTagsDelete(t *testing.T) {
	t.Run("deletes all managed tags", func(t *testing.T) {
		var deletes []string
		mock := &stackTagsClientMock{
			deleteStackTagFunc: func(_ context.Context, _ pulumiapi.StackIdentifier, tagName string) error {
				deletes = append(deletes, tagName)
				return nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)

		_, err := (&StackTags{}).Delete(ctx, infer.DeleteRequest[StackTagsState]{
			State: StackTagsState{
				Organization: "org",
				Project:      "project",
				Stack:        "stack",
				Tags:         map[string]string{"a": "1", "b": "2"},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, deletes)
	})

	t.Run("returns error if delete fails", func(t *testing.T) {
		mock := &stackTagsClientMock{
			deleteStackTagFunc: func(context.Context, pulumiapi.StackIdentifier, string) error {
				return errors.New("boom")
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)

		_, err := (&StackTags{}).Delete(ctx, infer.DeleteRequest[StackTagsState]{
			State: StackTagsState{
				Organization: "org",
				Project:      "project",
				Stack:        "stack",
				Tags:         map[string]string{"a": "1"},
			},
		})
		require.Error(t, err)
	})
}

func TestStackTagsRead(t *testing.T) {
	t.Run("returns only managed tags when prior inputs exist", func(t *testing.T) {
		mock := &stackTagsClientMock{
			getStackTagsFunc: func(context.Context, pulumiapi.StackIdentifier) (map[string]string, error) {
				return map[string]string{"tag1": "v1", "tag2": "v2", "extra": "x"}, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)

		resp, err := (&StackTags{}).Read(ctx, infer.ReadRequest[StackTagsInput, StackTagsState]{
			ID: "org/project/stack/tags",
			Inputs: StackTagsInput{
				Organization: "org",
				Project:      "project",
				Stack:        "stack",
				Tags:         map[string]string{"tag1": "old1", "tag2": "old2"},
			},
		})
		require.NoError(t, err)
		expected := StackTagsState{
			Organization: "org",
			Project:      "project",
			Stack:        "stack",
			Tags:         map[string]string{"tag1": "v1", "tag2": "v2"},
		}
		assert.Equal(t, infer.ReadResponse[StackTagsInput, StackTagsState]{
			ID:     "org/project/stack/tags",
			Inputs: expected,
			State:  expected,
		}, resp)
	})

	t.Run("import adopts every tag when there are no prior inputs", func(t *testing.T) {
		mock := &stackTagsClientMock{
			getStackTagsFunc: func(context.Context, pulumiapi.StackIdentifier) (map[string]string, error) {
				return map[string]string{"a": "1", "b": "2"}, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)

		resp, err := (&StackTags{}).Read(ctx, infer.ReadRequest[StackTagsInput, StackTagsState]{
			ID: "org/project/stack/tags",
		})
		require.NoError(t, err)
		expected := StackTagsState{
			Organization: "org",
			Project:      "project",
			Stack:        "stack",
			Tags:         map[string]string{"a": "1", "b": "2"},
		}
		assert.Equal(t, infer.ReadResponse[StackTagsInput, StackTagsState]{
			ID:     "org/project/stack/tags",
			Inputs: expected,
			State:  expected,
		}, resp)
	})

	t.Run("invalid id returns error", func(t *testing.T) {
		_, err := (&StackTags{}).Read(context.Background(),
			infer.ReadRequest[StackTagsInput, StackTagsState]{ID: "bogus"})
		require.Error(t, err)
	})
}
