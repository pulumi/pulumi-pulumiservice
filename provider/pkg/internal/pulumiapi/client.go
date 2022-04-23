// Copyright 2016-2022, Pulumi Corporation.
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
package pulumiapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	c       *http.Client
	token   string
	baseurl url.URL
}

func NewClient(args ...string) *Client {
	c := &http.Client{Timeout: time.Minute}

	var token string

	var baseURL = url.URL{
		Scheme: "https",
		Host:   "api.pulumi.com",
		Path:   "/api/",
	}

	if len(args) == 0 {
		token = ""
	} else {
		token = args[0]
	}

	if len(args) == 2 {
		token = args[0]
		baseURL.Host = args[1]
	}

	return &Client{
		c:       c,
		token:   token,
		baseurl: baseURL,
	}
}

func (c *Client) createRequest(ctx context.Context, method, path string, reqBody interface{}) (*http.Request, error) {
	var reqBodyReader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize %+v: %w", reqBody, err)
		}
		reqBodyReader = bytes.NewBuffer(data)
	}
	endpoint := c.baseurl.ResolveReference(&url.URL{Path: path})

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), reqBodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// add default headers
	req.Header.Add("Accept", "application/vnd.pulumi+8")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "token "+c.token)
	return req, nil
}

func (c *Client) sendRequest(req *http.Request, resBody interface{}) (*http.Response, error) {
	res, err := c.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	if !ok(res.StatusCode) {
		// if we didn't get an 2XX status code, unmarshal the response as an errorResponse
		// and return an error
		var errRes errorResponse
		err = json.Unmarshal(body, &errRes)
		if err != nil {
			return res, fmt.Errorf("failed to parse response body, status code %d: %w", res.StatusCode, err)
		}
		return res, &errRes
	}
	if resBody != nil {
		err = json.Unmarshal(body, resBody)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response body: %w", err)
		}
	}
	return res, nil
}

// do execute an http request to the pulumi service at the configured url
// Marshals reqBody and resBody to/from JSON. Applies appropriate headers as well
// Returns statusCode or error. If err != nil, statusCode == 0
func (c *Client) do(ctx context.Context, method, path string, reqBody interface{}, resBody interface{}) (*http.Response, error) {
	req, err := c.createRequest(ctx, method, path, reqBody)
	if err != nil {
		return nil, err
	}
	return c.sendRequest(req, resBody)
}
