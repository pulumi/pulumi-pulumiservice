// Copyright 2025, Pulumi Corporation.  All rights reserved.

package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
)

type CloudClientExecutor func(*http.Request) (*http.Response, error)

type CloudClientInterceptor func(ctx context.Context, parameters any) (any, bool, error)

type CloudClient struct {
	BaseURL     string
	Executor    CloudClientExecutor
	Interceptor CloudClientInterceptor
}

func (p *CloudClient) createRequest(
	ctx context.Context,
	method string,
	routePattern string,
	pathParams map[string]any,
	queryParams map[string]any,
) (*http.Request, error) {
	u, err := url.Parse(p.BaseURL)
	if err != nil {
		return nil, err
	}
	return createRequest(ctx, method, u, routePattern, pathParams, queryParams, nil)
}

func (p *CloudClient) createRequestWithRawBody(
	ctx context.Context,
	method string,
	routePattern string,
	pathParams map[string]any,
	queryParams map[string]any,
	body []byte,
) (*http.Request, error) {
	u, err := url.Parse(p.BaseURL)
	if err != nil {
		return nil, err
	}

	return createRequest(ctx, method, u, routePattern, pathParams, queryParams, bytes.NewReader(body))
}

func (p *CloudClient) createRequestWithBody(
	ctx context.Context,
	method string,
	routePattern string,
	pathParams map[string]any,
	queryParams map[string]any,
	body any,
) (*http.Request, error) {
	bodyRaw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := p.createRequestWithRawBody(ctx, method, routePattern, pathParams, queryParams, bodyRaw)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// Helper to construct a proper http.Request based on path and query.
func createRequest(
	ctx context.Context,
	method string,
	u *url.URL,
	routePattern string,
	pathParams map[string]any,
	queryParams map[string]any,
	body io.Reader,
) (*http.Request, error) {
	if len(pathParams) > 0 {
		for k, v := range pathParams {
			routePattern = strings.ReplaceAll(routePattern, "{"+k+"}", url.PathEscape(fmt.Sprint(v)))
		}
	}
	rawPath := routePattern
	path, err := url.PathUnescape(rawPath)
	if err != nil {
		return nil, err
	}
	u.Path = path
	u.RawPath = rawPath

	if len(queryParams) > 0 {
		q := url.Values{}
		for k, v := range queryParams {
			if v == nil {
				continue
			}
			rv := reflect.ValueOf(v)
			if rv.Kind() != reflect.Pointer {
				return nil, fmt.Errorf("query parameter '%s' must be a pointer", k)
			}
			if rv.IsNil() {
				continue
			}

			elem := rv.Elem()
			switch elem.Kind() {
			case reflect.Slice, reflect.Array:
				parts := make([]string, elem.Len())
				for i := range elem.Len() {
					parts[i] = fmt.Sprint(elem.Index(i).Interface())
				}
				q.Set(k, strings.Join(parts, ","))
			default:
				q.Set(k, fmt.Sprint(elem.Interface()))
			}
		}
		u.RawQuery = q.Encode()
	}

	return http.NewRequestWithContext(ctx, method, u.String(), body)
}

func (p *CloudClient) invokeRaw(req *http.Request, headers []http.Header) (*http.Response, error) {
	for _, header := range headers {
		for key, vals := range header {
			for i, val := range vals {
				if i == 0 {
					req.Header.Set(key, val)
				} else {
					req.Header.Add(key, val)
				}
			}
		}
	}

	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/vnd.pulumi+8")
	}
	if strings.Contains(p.BaseURL, "ngrok") {
		// ngrok is used for local development, so we need to allow a bypass header
		// See https://ngrok.com/abuse
		req.Header.Set("ngrok-skip-browser-warning", "true")
	}

	resp, err := p.Executor(req)
	if err != nil {
		return nil, fmt.Errorf("performing HTTP request: %w", err)
	}

	// For 4xx and 5xx failures, attempt to provide better diagnostics about what may have gone wrong.
	if resp.StatusCode >= 400 && resp.StatusCode <= 599 {
		// 4xx and 5xx responses should be of type ErrorResponse. See if we can unmarshal as that
		// type, and if not just return the raw response text.
		respBody, err := io.ReadAll(resp.Body)
		contract.IgnoreClose(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("API call failed (%s), could not read response: %w", resp.Status, err)
		}

		var parsed struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err = json.Unmarshal(respBody, &parsed); err != nil {
			parsed.Code = resp.StatusCode
			parsed.Message = strings.TrimSpace(string(respBody))
		}

		if parsed.Code == 0 {
			// Our error parsed as JSON but it doesn't match the schema
			// returned by API. Use the HTTP status code and raw body.
			return nil, NewAPIError(resp.StatusCode, strings.TrimSpace(string(respBody)), resp.Header.Clone())
		}

		return nil, NewAPIError(parsed.Code, parsed.Message, resp.Header.Clone())
	}

	return resp, nil
}

func (p *CloudClient) invoke(req *http.Request, headers []http.Header) error {
	_, err := p.invokeRaw(req, headers)
	return err
}

func (p *CloudClient) invokeWithResponse(req *http.Request, headers []http.Header) ([]byte, error) {
	reader, err := p.invokeWithStreamingResponse(req, headers)
	if err != nil {
		return nil, err
	}

	defer reader.Close()

	return io.ReadAll(reader)
}

func (p *CloudClient) invokeWithStreamingResponse(req *http.Request, headers []http.Header) (io.ReadCloser, error) {
	resp, err := p.invokeRaw(req, headers)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNoContent {
		_ = resp.Body.Close()
		return nil, NewAPIError(http.StatusNoContent, "API call returned 204 No Content", resp.Header.Clone())
	}

	return resp.Body, nil
}
