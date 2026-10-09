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

package functions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type stackClientMock struct {
	config.Client
	stack *apitype.AppStack
}

func (c *stackClientMock) GetStack(_ context.Context, _ pulumiapi.StackIdentifier) (*apitype.AppStack, error) {
	return c.stack, nil
}

const testStackID = "stack-uuid"

func TestGetStack(t *testing.T) {
	t.Parallel()

	in := GetStackInput{OrganizationName: testOrgName, ProjectName: "proj", StackName: "prod"}
	invoke := func(t *testing.T, s *apitype.AppStack) (GetStackOutput, error) {
		ctx := config.WithMockClient(t.Context(), &stackClientMock{stack: s})
		resp, err := GetStackFunction{}.Invoke(ctx, infer.FunctionRequest[GetStackInput]{Input: in})
		return resp.Output, err
	}

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		got, err := invoke(t, &apitype.AppStack{ID: testStackID, Tags: map[string]string{"team": "platform"}})
		require.NoError(t, err)
		assert.Equal(t, GetStackOutput{
			OrganizationName: testOrgName,
			ProjectName:      "proj",
			StackName:        "prod",
			StackId:          testStackID,
			Tags:             map[string]string{"team": "platform"},
		}, got)
	})

	t.Run("no tags", func(t *testing.T) {
		t.Parallel()
		got, err := invoke(t, &apitype.AppStack{ID: testStackID})
		require.NoError(t, err)
		assert.Equal(t, map[string]string{}, got.Tags)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, nil)
		require.ErrorContains(t, err, "stack org/proj/prod not found")
	})
}
