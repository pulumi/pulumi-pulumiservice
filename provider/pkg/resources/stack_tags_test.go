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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type StackTagClientMock struct {
	config.Client
	getStackTagFunc func(
		ctx context.Context, stackName pulumiapi.StackIdentifier, tagName string,
	) (*pulumiapi.StackTag, error)
}

func (c *StackTagClientMock) GetStackTag(
	ctx context.Context, stackName pulumiapi.StackIdentifier, tagName string,
) (*pulumiapi.StackTag, error) {
	return c.getStackTagFunc(ctx, stackName, tagName)
}

func TestStackTag(t *testing.T) {
	t.Parallel()

	t.Run("Read when the resource is not found", func(t *testing.T) {
		t.Parallel()
		mockedClient := &StackTagClientMock{
			getStackTagFunc: func(_ context.Context, _ pulumiapi.StackIdentifier, _ string) (*pulumiapi.StackTag, error) {
				return nil, nil
			},
		}

		ctx := config.WithMockClient(context.Background(), mockedClient)

		st := &StackTag{}
		req := infer.ReadRequest[StackTagInput, StackTagState]{
			ID: "org/project/stack/tag",
			Inputs: StackTagInput{
				Organization: "org",
				Project:      "project",
				Stack:        "stack",
				Name:         "tag",
				Value:        "value",
			},
			State: StackTagState{
				Organization: "org",
				Project:      "project",
				Stack:        "stack",
				Name:         "tag",
				Value:        "value",
			},
		}

		resp, err := st.Read(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, infer.ReadResponse[StackTagInput, StackTagState]{}, resp)
	})

	t.Run("Read when the resource is found", func(t *testing.T) {
		t.Parallel()
		mockedClient := &StackTagClientMock{
			getStackTagFunc: func(_ context.Context, _ pulumiapi.StackIdentifier, _ string) (*pulumiapi.StackTag, error) {
				return &pulumiapi.StackTag{
					Name:  "myTag",
					Value: "myValue",
				}, nil
			},
		}

		ctx := config.WithMockClient(context.Background(), mockedClient)

		st := &StackTag{}
		req := infer.ReadRequest[StackTagInput, StackTagState]{
			ID: "org/project/stack/myTag",
			Inputs: StackTagInput{
				Organization: "org",
				Project:      "project",
				Stack:        "stack",
				Name:         "myTag",
				Value:        "oldValue",
			},
			State: StackTagState{
				Organization: "org",
				Project:      "project",
				Stack:        "stack",
				Name:         "myTag",
				Value:        "oldValue",
			},
		}

		resp, err := st.Read(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, infer.ReadResponse[StackTagInput, StackTagState]{
			ID: "org/project/stack/myTag",
			Inputs: StackTagInput{
				Organization: "org",
				Project:      "project",
				Stack:        "stack",
				Name:         "myTag",
				Value:        "myValue",
			},
			State: StackTagState{
				Organization: "org",
				Project:      "project",
				Stack:        "stack",
				Name:         "myTag",
				Value:        "myValue",
			},
		}, resp)
	})
}
