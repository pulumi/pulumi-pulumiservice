package pulumiapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testServerConfig struct {
	ExpectedReqMethod   string
	ExpectedReqBody     interface{}
	ExpectedReqPath     string
	ExpectedQueryParams url.Values
	ResponseBody        interface{}
	ResponseCode        int
}

func startTestServer(t *testing.T, config testServerConfig) (client *Client, cleanup func()) {
	token := "abc123"
	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// ensure that proper http verb was used as well as the path
		assert.Equal(t, config.ExpectedReqMethod, r.Method)
		assert.Equal(t, config.ExpectedReqPath, r.URL.Path)
		if config.ExpectedQueryParams != nil {
			assert.Equal(t, config.ExpectedQueryParams, r.URL.Query())
		}

		// these should always be set, so always test for them
		assert.Equal(t, "token "+token, r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/vnd.pulumi+8", r.Header.Get("Accept"))
		assert.Equal(t, "provider", r.Header.Get("X-Pulumi-Source"))

		// if we expected a request body, unmarshal the body and
		if config.ExpectedReqBody != nil {
			expectedBody, _ := json.Marshal(config.ExpectedReqBody)
			actualBody, _ := io.ReadAll(r.Body)
			assert.JSONEq(t, string(expectedBody), string(actualBody))
		}
		w.WriteHeader(config.ResponseCode)
		if config.ResponseBody != nil {
			resBytes, _ := json.Marshal(config.ResponseBody)
			_, _ = w.Write(resBytes)
		}
	}))
	c, err := NewClient(&httpClient, token, server.URL)
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}
	return c, server.Close
}
