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

package pulumiapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSPOrg     = "an-org"
	testSPLogin   = "service-principal:d333a711"
	testSPName    = "ci-deployer"
	testSPTokenID = "tok-123"
	testSPPath    = "/api/orgs/an-org/service-principals/service-principal:d333a711"
)

func spRolesResponse(ids ...string) map[string]any {
	roles := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		roles = append(roles, map[string]any{"id": id, "name": "role " + id})
	}
	return map[string]any{"roles": roles}
}

func TestCreateServicePrincipal(t *testing.T) {
	t.Run("sends name and roles, returns the login", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/an-org/service-principals/create",
			ExpectedReqBody:   map[string]any{"name": testSPName, "roleIds": []string{"role-1"}},
			ResponseCode:      http.StatusCreated,
			ResponseBody: map[string]any{
				"role": "member",
				"user": map[string]any{
					"name":                 testSPName,
					"githubLogin":          testSPLogin,
					"servicePrincipalKind": "user-defined",
				},
			},
		})
		sp, err := c.CreateServicePrincipal(t.Context(), testSPOrg, testSPName, []string{"role-1"})
		require.NoError(t, err)
		assert.Equal(t, &ServicePrincipal{Login: testSPLogin, Name: testSPName, RoleIDs: []string{"role-1"}}, sp)
	})

	t.Run("omits roles when there are none", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/an-org/service-principals/create",
			ExpectedReqBody:   map[string]any{"name": testSPName},
			ResponseCode:      http.StatusCreated,
			ResponseBody:      map[string]any{"user": map[string]any{"name": testSPName, "githubLogin": testSPLogin}},
		})
		sp, err := c.CreateServicePrincipal(t.Context(), testSPOrg, testSPName, nil)
		require.NoError(t, err)
		assert.Equal(t, testSPLogin, sp.Login)
	})

	t.Run("surfaces the service's error", func(t *testing.T) {
		msg := "without custom roles, a service principal holds at most one of the Reader, Maintainer or Admin roles"
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/an-org/service-principals/create",
			ResponseCode:      http.StatusBadRequest,
			ResponseBody:      ErrorResponse{StatusCode: http.StatusBadRequest, Message: msg},
		})
		_, err := c.CreateServicePrincipal(t.Context(), testSPOrg, testSPName, []string{"a", "b"})
		assert.ErrorContains(t, err, msg)
		assert.Equal(t, http.StatusBadRequest, GetErrorStatusCode(err))
	})

	t.Run("rejects empty arguments", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{})
		_, err := c.CreateServicePrincipal(t.Context(), "", testSPName, nil)
		assert.Error(t, err)
		_, err = c.CreateServicePrincipal(t.Context(), testSPOrg, "", nil)
		assert.Error(t, err)
	})
}

func TestGetServicePrincipal(t *testing.T) {
	members := func(logins ...string) map[string]any {
		var ms []map[string]any
		for _, l := range logins {
			ms = append(ms, map[string]any{"user": map[string]any{"name": "name of " + l, "githubLogin": l}})
		}
		return map[string]any{"members": ms}
	}

	t.Run("reads the name from the member list and the roles", func(t *testing.T) {
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			switch {
			case r.Method == http.MethodGet && r.URL.Path == "/api/orgs/an-org/members":
				assert.Equal(t, "service-principal", r.URL.Query().Get("type"))
				return http.StatusOK, members("service-principal:other", testSPLogin)
			case r.Method == http.MethodGet && r.URL.Path == testSPPath+"/roles":
				return http.StatusOK, spRolesResponse("role-1", "role-2")
			}
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			return http.StatusInternalServerError, nil
		})
		sp, err := c.GetServicePrincipal(t.Context(), testSPOrg, testSPLogin)
		require.NoError(t, err)
		assert.Equal(t, &ServicePrincipal{
			Login:   testSPLogin,
			Name:    "name of " + testSPLogin,
			RoleIDs: []string{"role-1", "role-2"},
		}, sp)
	})

	t.Run("follows the member list's continuation token", func(t *testing.T) {
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			if r.URL.Path == testSPPath+"/roles" {
				return http.StatusOK, spRolesResponse()
			}
			if r.URL.Query().Get("continuationToken") == "" {
				page := members("service-principal:other")
				page["continuationToken"] = "next"
				return http.StatusOK, page
			}
			return http.StatusOK, members(testSPLogin)
		})
		sp, err := c.GetServicePrincipal(t.Context(), testSPOrg, testSPLogin)
		require.NoError(t, err)
		require.NotNil(t, sp)
		assert.Empty(t, sp.RoleIDs)
	})

	t.Run("missing from the member list", func(t *testing.T) {
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			assert.Equal(t, "/api/orgs/an-org/members", r.URL.Path, "roles must not be read")
			return http.StatusOK, members("service-principal:other")
		})
		sp, err := c.GetServicePrincipal(t.Context(), testSPOrg, testSPLogin)
		require.NoError(t, err)
		assert.Nil(t, sp)
	})

	t.Run("roles not found", func(t *testing.T) {
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			if r.URL.Path == testSPPath+"/roles" {
				return http.StatusNotFound, ErrorResponse{StatusCode: http.StatusNotFound, Message: "not found"}
			}
			return http.StatusOK, members(testSPLogin)
		})
		sp, err := c.GetServicePrincipal(t.Context(), testSPOrg, testSPLogin)
		require.NoError(t, err)
		assert.Nil(t, sp)
	})
}

func TestRenameServicePrincipal(t *testing.T) {
	c := startTestServer(t, testServerConfig{
		ExpectedReqMethod: http.MethodPatch,
		ExpectedReqPath:   "/api/orgs/an-org/members/" + testSPLogin,
		ExpectedReqBody:   map[string]any{"name": "new-name"},
		ResponseCode:      http.StatusNoContent,
	})
	require.NoError(t, c.RenameServicePrincipal(t.Context(), testSPOrg, testSPLogin, "new-name"))
	assert.Error(t, c.RenameServicePrincipal(t.Context(), testSPOrg, testSPLogin, ""))
}

func TestUpdateServicePrincipalRoles(t *testing.T) {
	t.Run("sets the roles", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPut,
			ExpectedReqPath:   testSPPath + "/roles",
			ExpectedReqBody:   map[string]any{"roleIds": []string{"role-1"}},
			ResponseCode:      http.StatusOK,
			ResponseBody:      spRolesResponse("role-1"),
		})
		ids, err := c.UpdateServicePrincipalRoles(t.Context(), testSPOrg, testSPLogin, []string{"role-1"})
		require.NoError(t, err)
		assert.Equal(t, []string{"role-1"}, ids)
	})

	t.Run("no roles sends an empty list", func(t *testing.T) {
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			assert.JSONEq(t, `{"roleIds":[]}`, string(body))
			return http.StatusOK, spRolesResponse()
		})
		ids, err := c.UpdateServicePrincipalRoles(t.Context(), testSPOrg, testSPLogin, nil)
		require.NoError(t, err)
		assert.Empty(t, ids)
	})
}

func TestCreateServicePrincipalToken(t *testing.T) {
	expires := time.Date(2027, 6, 30, 17, 0, 0, 0, time.UTC)

	t.Run("sends the expiry as a date-time", func(t *testing.T) {
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, testSPPath+"/tokens", r.URL.Path)
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			assert.JSONEq(t,
				`{"name":"ci","description":"for CI","expires":"2027-06-30T17:00:00Z"}`, string(body))
			return http.StatusOK, createTokenResponse{ID: testSPTokenID, TokenValue: "secret"}
		})
		token, err := c.CreateServicePrincipalToken(t.Context(), testSPOrg, testSPLogin,
			CreateServicePrincipalTokenRequest{Name: "ci", Description: "for CI", Expires: &expires})
		require.NoError(t, err)
		assert.Equal(t, &AccessToken{
			ID:          testSPTokenID,
			Name:        "ci",
			TokenValue:  "secret",
			Description: "for CI",
			Expires:     expires.Unix(),
		}, token)
	})

	t.Run("omits a missing expiry", func(t *testing.T) {
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.NotContains(t, body, "expires")
			assert.Contains(t, body, "description")
			return http.StatusOK, createTokenResponse{ID: testSPTokenID, TokenValue: "secret"}
		})
		_, err := c.CreateServicePrincipalToken(t.Context(), testSPOrg, testSPLogin,
			CreateServicePrincipalTokenRequest{Name: "ci"})
		require.NoError(t, err)
	})

	t.Run("rejects empty arguments", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{})
		_, err := c.CreateServicePrincipalToken(t.Context(), testSPOrg, "", CreateServicePrincipalTokenRequest{Name: "ci"})
		assert.Error(t, err)
		_, err = c.CreateServicePrincipalToken(t.Context(), testSPOrg, testSPLogin, CreateServicePrincipalTokenRequest{})
		assert.Error(t, err)
	})
}

func TestGetServicePrincipalToken(t *testing.T) {
	token := func(id string) accessTokenResponse {
		return accessTokenResponse{ID: id, Name: "name-" + id, Description: "desc", Expires: 1814374800}
	}

	t.Run("finds the token among all tokens, expired ones included", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod:   http.MethodGet,
			ExpectedReqPath:     testSPPath + "/tokens",
			ExpectedQueryParams: url.Values{"filter": {"all"}, "search": {testSPTokenID}},
			ResponseCode:        http.StatusOK,
			ResponseBody:        listTokenResponse{Tokens: []accessTokenResponse{token("other"), token(testSPTokenID)}},
		})
		got, err := c.GetServicePrincipalToken(t.Context(), testSPOrg, testSPLogin, testSPTokenID)
		require.NoError(t, err)
		assert.Equal(t, &AccessToken{
			ID:          testSPTokenID,
			Name:        "name-" + testSPTokenID,
			Description: "desc",
			Expires:     1814374800,
		}, got)
	})

	t.Run("follows the continuation token", func(t *testing.T) {
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			if r.URL.Query().Get("continuationToken") == "" {
				return http.StatusOK, map[string]any{
					"tokens":            []accessTokenResponse{token("other")},
					"continuationToken": "next",
				}
			}
			assert.Equal(t, "next", r.URL.Query().Get("continuationToken"))
			return http.StatusOK, listTokenResponse{Tokens: []accessTokenResponse{token(testSPTokenID)}}
		})
		got, err := c.GetServicePrincipalToken(t.Context(), testSPOrg, testSPLogin, testSPTokenID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, testSPTokenID, got.ID)
	})

	t.Run("token not listed", func(t *testing.T) {
		c := startTestServerMulti(t, func(*http.Request) (int, any) {
			return http.StatusOK, listTokenResponse{Tokens: []accessTokenResponse{token("other")}}
		})
		got, err := c.GetServicePrincipalToken(t.Context(), testSPOrg, testSPLogin, testSPTokenID)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("service principal not found", func(t *testing.T) {
		c := startTestServerMulti(t, func(*http.Request) (int, any) {
			return http.StatusNotFound, ErrorResponse{StatusCode: http.StatusNotFound, Message: "not found"}
		})
		got, err := c.GetServicePrincipalToken(t.Context(), testSPOrg, testSPLogin, testSPTokenID)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("server error", func(t *testing.T) {
		c := startTestServerMulti(t, func(*http.Request) (int, any) {
			return http.StatusInternalServerError, ErrorResponse{StatusCode: 500, Message: "boom"}
		})
		_, err := c.GetServicePrincipalToken(t.Context(), testSPOrg, testSPLogin, testSPTokenID)
		assert.Error(t, err)
	})
}
