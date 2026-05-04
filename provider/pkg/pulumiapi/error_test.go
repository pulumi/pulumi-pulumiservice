// Copyright 2016-2026, Pulumi Corporation.
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
package pulumiapi

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/apiclient"
)

func TestGetErrorStatusCode(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		assert.Equal(t, 0, GetErrorStatusCode(nil))
	})

	t.Run("non-API error", func(t *testing.T) {
		assert.Equal(t, 0, GetErrorStatusCode(errors.New("plain error")))
	})

	t.Run("legacy ErrorResponse", func(t *testing.T) {
		err := &ErrorResponse{StatusCode: http.StatusConflict, Message: "in use"}
		assert.Equal(t, http.StatusConflict, GetErrorStatusCode(err))
	})

	t.Run("legacy ErrorResponse wrapped", func(t *testing.T) {
		base := &ErrorResponse{StatusCode: http.StatusNotFound}
		wrapped := fmt.Errorf("outer: %w", base)
		assert.Equal(t, http.StatusNotFound, GetErrorStatusCode(wrapped))
	})

	t.Run("SDK APIError", func(t *testing.T) {
		err := apiclient.NewAPIError(http.StatusConflict, "in use", nil)
		assert.Equal(t, http.StatusConflict, GetErrorStatusCode(err))
	})

	t.Run("SDK APIError wrapped", func(t *testing.T) {
		base := apiclient.NewAPIError(http.StatusNotFound, "missing", nil)
		wrapped := fmt.Errorf("outer: %w", base)
		assert.Equal(t, http.StatusNotFound, GetErrorStatusCode(wrapped))
	})
}
