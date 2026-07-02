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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

const policyPacksPath = "/api/orgs/anOrg/policypacks"

func TestListPolicyPacks(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		want := []PolicyPackWithVersions{
			{
				Name:        alphaPolicyPack,
				DisplayName: "Alpha",
				Versions:    []int{1, 2},
				VersionTags: []string{policyPackVersion, "1.1.0"},
			},
		}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   policyPacksPath,
			ResponseCode:      http.StatusOK,
			ResponseBody:      listPolicyPacksResponse{PolicyPacks: want},
		})
		got, err := c.ListPolicyPacks(ctx, "anOrg")
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("empty org rejected", func(t *testing.T) {
		c := &Client{}
		_, err := c.ListPolicyPacks(ctx, "")
		assert.EqualError(t, err, "empty orgName")
	})
}

func TestGetPolicyPack(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		want := &PolicyPackDetail{Name: alphaPolicyPack, Version: 2, VersionTag: "1.1.0"}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/anOrg/policypacks/alpha/versions/2",
			ResponseCode:      http.StatusOK,
			ResponseBody:      want,
		})
		got, err := c.GetPolicyPack(ctx, "anOrg", alphaPolicyPack, 2)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("404 returns nil pack and nil error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/anOrg/policypacks/missing/versions/1",
			ResponseCode:      http.StatusNotFound,
			ResponseBody:      ErrorResponse{StatusCode: 404, Message: notFoundError},
		})
		got, err := c.GetPolicyPack(ctx, "anOrg", "missing", 1)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("input validation", func(t *testing.T) {
		c := &Client{}
		_, err := c.GetPolicyPack(ctx, "", alphaPolicyPack, 1)
		assert.EqualError(t, err, "empty orgName")
		_, err = c.GetPolicyPack(ctx, "anOrg", "", 1)
		assert.EqualError(t, err, "empty policyPackName")
	})
}

func TestGetLatestPolicyPack(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		want := &PolicyPackDetail{Name: alphaPolicyPack, Version: 9, VersionTag: "9.0.0"}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/anOrg/policypacks/alpha/latest",
			ResponseCode:      http.StatusOK,
			ResponseBody:      want,
		})
		got, err := c.GetLatestPolicyPack(ctx, "anOrg", alphaPolicyPack)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("404 returns nil pack and nil error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/anOrg/policypacks/missing/latest",
			ResponseCode:      http.StatusNotFound,
			ResponseBody:      ErrorResponse{StatusCode: 404, Message: notFoundError},
		})
		got, err := c.GetLatestPolicyPack(ctx, "anOrg", "missing")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestDeletePolicyPackVersion(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   policyPackVersionPath,
			ResponseCode:      http.StatusNoContent,
		})
		assert.NoError(t, c.DeletePolicyPackVersion(ctx, "anOrg", alphaPolicyPack, policyPackVersion))
	})

	t.Run("404 is a no-op", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   policyPackVersionPath,
			ResponseCode:      http.StatusNotFound,
			ResponseBody:      ErrorResponse{StatusCode: 404, Message: notFoundError},
		})
		assert.NoError(t, c.DeletePolicyPackVersion(ctx, "anOrg", alphaPolicyPack, policyPackVersion))
	})

	t.Run("input validation", func(t *testing.T) {
		c := &Client{}
		assert.EqualError(t, c.DeletePolicyPackVersion(ctx, "", "p", "v"), "empty orgName")
		assert.EqualError(t, c.DeletePolicyPackVersion(ctx, "o", "", "v"), "empty policy pack name")
		assert.EqualError(t, c.DeletePolicyPackVersion(ctx, "o", "p", ""), "empty versionTag")
	})
}

func TestDeletePolicyPack(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/anOrg/policypacks/alpha",
			ResponseCode:      http.StatusNoContent,
		})
		assert.NoError(t, c.DeletePolicyPack(ctx, "anOrg", alphaPolicyPack))
	})

	t.Run("404 is a no-op", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/anOrg/policypacks/alpha",
			ResponseCode:      http.StatusNotFound,
			ResponseBody:      ErrorResponse{StatusCode: 404, Message: notFoundError},
		})
		assert.NoError(t, c.DeletePolicyPack(ctx, "anOrg", alphaPolicyPack))
	})

	t.Run("input validation", func(t *testing.T) {
		c := &Client{}
		assert.EqualError(t, c.DeletePolicyPack(ctx, "", "p"), "empty orgName")
		assert.EqualError(t, c.DeletePolicyPack(ctx, "o", ""), "empty policy pack name")
	})
}

func TestPublishPolicyPack_HappyPath(t *testing.T) {
	archive := []byte("tarball-bytes")
	uploadHit := false
	completeHit := false

	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uploadHit = true
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, valueKey, r.Header.Get("X-Required"))
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, archive, body)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(uploadServer.Close)

	c := startTestServerMulti(t, func(r *http.Request) (int, any) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == policyPacksPath:
			var got CreatePolicyPackRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
			assert.Equal(t, alphaPolicyPack, got.Name)
			assert.Equal(t, policyPackVersion, got.VersionTag)
			return http.StatusOK, CreatePolicyPackResponse{
				Version:         7,
				UploadURI:       uploadServer.URL + "/upload",
				RequiredHeaders: map[string]string{"X-Required": valueKey},
			}
		case r.Method == http.MethodPost && r.URL.Path == "/api/orgs/anOrg/policypacks/alpha/versions/1.0.0/complete":
			completeHit = true
			return http.StatusOK, nil
		}
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		return http.StatusInternalServerError, nil
	})

	version, err := c.PublishPolicyPack(ctx, "anOrg", CreatePolicyPackRequest{
		Name:       alphaPolicyPack,
		VersionTag: policyPackVersion,
		Policies:   []apitype.Policy{{Name: "rule"}},
	}, bytes.NewReader(archive))
	require.NoError(t, err)
	assert.Equal(t, 7, version)
	assert.True(t, uploadHit, "upload should have been called")
	assert.True(t, completeHit, "complete should have been called")
}

func TestPublishPolicyPack_UploadFailureTriggersCleanup(t *testing.T) {
	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad blob"))
	}))
	t.Cleanup(uploadServer.Close)

	cleanupHit := false
	c := startTestServerMulti(t, func(r *http.Request) (int, any) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == policyPacksPath:
			return http.StatusOK, CreatePolicyPackResponse{
				Version:   7,
				UploadURI: uploadServer.URL + "/upload",
			}
		case r.Method == http.MethodDelete && r.URL.Path == policyPackVersionPath:
			cleanupHit = true
			return http.StatusNoContent, nil
		}
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		return http.StatusInternalServerError, nil
	})

	_, err := c.PublishPolicyPack(ctx, "anOrg", CreatePolicyPackRequest{
		Name:       alphaPolicyPack,
		VersionTag: policyPackVersion,
	}, strings.NewReader("payload"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "upload failed")
	assert.True(t, cleanupHit, "orphaned version should be cleaned up on failure")
}

func TestPublishPolicyPack_CleanupUsesDetachedContext(t *testing.T) {
	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(uploadServer.Close)

	canceledCtx, cancel := context.WithCancel(ctx)
	cleanupHit := false
	c := startTestServerMulti(t, func(r *http.Request) (int, any) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == policyPacksPath:
			return http.StatusOK, CreatePolicyPackResponse{
				Version:   7,
				UploadURI: uploadServer.URL + "/upload",
			}
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/complete"):
			// cancel the caller's context just before completion to make complete fail
			cancel()
			return http.StatusInternalServerError, ErrorResponse{StatusCode: 500, Message: "complete failed"}
		case r.Method == http.MethodDelete && r.URL.Path == policyPackVersionPath:
			cleanupHit = true
			return http.StatusNoContent, nil
		}
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		return http.StatusInternalServerError, nil
	})

	_, err := c.PublishPolicyPack(canceledCtx, "anOrg", CreatePolicyPackRequest{
		Name:       alphaPolicyPack,
		VersionTag: policyPackVersion,
	}, strings.NewReader("payload"))
	require.Error(t, err)
	assert.True(t, cleanupHit, "cleanup should run even when caller context is canceled")
}

func TestUploadToSignedURL_Retries5xx(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, valueKey, r.Header.Get("X-Required"))
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, []byte("payload"), body)
		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	c, err := NewClient(&http.Client{}, "tok", server.URL)
	require.NoError(t, err)

	err = c.uploadToSignedURL(ctx, server.URL+"/upload",
		map[string]string{"X-Required": valueKey}, []byte("payload"))
	require.NoError(t, err)
	assert.EqualValues(t, 2, attempts.Load(), "expected one retry after the first 5xx")
}

func TestUploadToSignedURL_DoesNotRetry4xx(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("nope"))
	}))
	t.Cleanup(server.Close)

	c, err := NewClient(&http.Client{}, "tok", server.URL)
	require.NoError(t, err)

	err = c.uploadToSignedURL(ctx, server.URL+"/upload", nil, []byte("payload"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "upload failed")
	assert.EqualValues(t, 1, attempts.Load(), "4xx should not retry")
}

func TestPublishPolicyPack_InputValidation(t *testing.T) {
	c := &Client{}
	_, err := c.PublishPolicyPack(ctx, "", CreatePolicyPackRequest{Name: "a", VersionTag: "1"}, nil)
	assert.EqualError(t, err, "empty orgName")
	_, err = c.PublishPolicyPack(ctx, "o", CreatePolicyPackRequest{VersionTag: "1"}, nil)
	assert.EqualError(t, err, "empty policy pack name")
	_, err = c.PublishPolicyPack(ctx, "o", CreatePolicyPackRequest{Name: "a"}, nil)
	assert.EqualError(t, err, "empty versionTag")
}
