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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/apiclient"
)

type Client struct {
	httpClient *http.Client
	token      string
	baseurl    *url.URL
	SDK        *apiclient.CloudClient
}

func NewClient(client *http.Client, token, URL string) (*Client, error) {

	var baseURL = &url.URL{
		Scheme: "https",
		Host:   "api.pulumi.com",
		Path:   "/api/",
	}
	if len(URL) > 0 {
		parsedURL, err := url.Parse(URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL (%q): %w", URL, err)
		}
		baseURL = parsedURL
		baseURL.Path = "/api/"
	}

	sendRequest := func(req *http.Request) (*http.Response, error) {
		// Match the headers the hand-rolled c.do() flow attaches in
		// createRequest below — Accept defaults to vnd.pulumi+8 (let the
		// SDK's invokeRaw default win; do not override here),
		// X-Pulumi-Source attributes provider writes for Cloud audit, and
		// Content-Type is set unconditionally to mirror the legacy path.
		if req.Header.Get("X-Pulumi-Source") == "" {
			req.Header.Set("X-Pulumi-Source", "provider")
		}
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
		req.Header.Set("User-Agent", "pulumi-admin/1")

		return client.Do(req) //nolint:gosec // G704 — URL is from trusted admin config
	}

	pulumiAPI := &apiclient.CloudClient{
		BaseURL:  baseURL.String(),
		Executor: sendRequest,
	}

	return &Client{
		SDK:        pulumiAPI,
		httpClient: client,
		token:      token,
		baseurl:    baseURL,
	}, nil
}

// createRequest creates a *http.Request with standard headers set and reqBody marshalled into json.
func (c *Client) createRequest(
	ctx context.Context,
	method string,
	url *url.URL,
	reqBody interface{},
) (*http.Request, error) {
	var reqBodyReader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize %+v: %w", reqBody, err)
		}
		reqBodyReader = bytes.NewBuffer(data)
	}
	endpoint := c.baseurl.ResolveReference(url)

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), reqBodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// add default headers
	req.Header.Add("X-Pulumi-Source", "provider")
	req.Header.Add("Accept", "application/vnd.pulumi+8")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "token "+c.token)

	return req, nil
}

// sendRequest executes req and unmarshals response json into resBody
// returns attempts to unmarshal response into ErrorResponse if statusCode not 2XX
func (c *Client) sendRequest(req *http.Request, resBody interface{}) (*http.Response, error) {
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	defer func() {
		err = res.Body.Close()
		if err != nil {
			fmt.Println("failed to close reading body: %w", err)
		}
	}()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	if !ok(res.StatusCode) {
		// if we didn't get an 2XX status code, unmarshal the response as an ErrorResponse
		// and return an error. Some edges (e.g. gorilla mux 404) return
		// plain-text bodies, so fall back to the raw body as the message
		// rather than masking the real status with a JSON parse error.
		var errRes ErrorResponse
		if unmarshalErr := json.Unmarshal(body, &errRes); unmarshalErr != nil {
			errRes = ErrorResponse{Message: strings.TrimSpace(string(body))}
		}
		if errRes.StatusCode == 0 {
			errRes.StatusCode = res.StatusCode
		}
		return res, &errRes
	}
	// Only unmarshal response body if:
	// 1. resBody is not nil (caller wants the response)
	// 2. Body is not empty (HTTP 204 No Content returns empty body)
	if resBody != nil && len(body) > 0 {
		err = json.Unmarshal(body, resBody)
		if err != nil {
			return res, fmt.Errorf("failed to parse response body: %w", err)
		}
	}
	return res, nil
}

// do execute an http request to the pulumi service at the configured url
// Marshals reqBody and resBody to/from JSON. Applies appropriate headers as well
// Returns http.Response, but Body will be closed
func (c *Client) do(
	ctx context.Context,
	method, path string,
	reqBody interface{},
	resBody interface{},
) (*http.Response, error) {
	req, err := c.createRequest(ctx, method, &url.URL{Path: path}, reqBody)
	if err != nil {
		return nil, err
	}
	return c.sendRequest(req, resBody)
}

func (c *Client) doWithQuery(
	ctx context.Context,
	method, path string,
	query url.Values,
	reqBody interface{},
	resBody interface{},
) (*http.Response, error) {
	reqURL := &url.URL{Path: path}
	reqURL.RawQuery = query.Encode()
	req, err := c.createRequest(ctx, method, reqURL, reqBody)
	if err != nil {
		return nil, err
	}
	return c.sendRequest(req, resBody)
}
