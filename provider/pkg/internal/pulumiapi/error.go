// Copyright 2016-2022, Pulumi Corporation.
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
)

// ErrorResponse is returned from pulumi service api when there's been an error
type ErrorResponse struct {
	StatusCode int    `json:"code"`
	Message    string `json:"message"`
}

func (err *ErrorResponse) Error() string {
	return fmt.Sprintf("%d API error: %s", err.StatusCode, err.Message)
}

func GetErrorStatusCode(err error) int {
	var errResp *ErrorResponse
	if errors.As(err, &errResp) {
		return errResp.StatusCode
	}
	return 0
}
