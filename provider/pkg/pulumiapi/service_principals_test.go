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
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSPOrg    = "an-org"
	testSPLogin  = "service-principal:d333a711"
	testSPName   = "ci-deployer"
	testSPPath   = "/api/orgs/an-org/service-principals/service-principal:d333a711"
	testSPCreate = "/api/orgs/an-org/service-principals/create"
	testSPRoleA  = "sp-role-a"
	testSPRoleB  = "sp-role-b"
)

func spRolesResponse(ids ...string) servicePrincipalRolesResponse {
	res := servicePrincipalRolesResponse{Roles: []servicePrincipalRole{}}
	for _, id := range ids {
		res.Roles = append(res.Roles, servicePrincipalRole{ID: id})
	}
	return res
}

func spMember(name, login string) servicePrincipalMember {
	return servicePrincipalMember{User: User{Name: name, GithubLogin: login}}
}

func spMembers(logins ...string) Members {
	var page Members
	for _, l := range logins {
		page.Members = append(page.Members, Member{User: User{Name: "name of " + l, GithubLogin: l}})
	}
	return page
}

func TestCreateServicePrincipal(t *testing.T) {
	t.Run("sends name and roles, returns the login", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   testSPCreate,
			ExpectedReqBody:   createServicePrincipalRequest{Name: testSPName, RoleIDs: []string{testSPRoleA}},
			ResponseCode:      http.StatusCreated,
			ResponseBody:      spMember(testSPName, testSPLogin),
		})
		sp, err := c.CreateServicePrincipal(t.Context(), testSPOrg, testSPName, []string{testSPRoleA})
		require.NoError(t, err)
		assert.Equal(t, &ServicePrincipal{Login: testSPLogin, Name: testSPName, RoleIDs: []string{testSPRoleA}}, sp)
	})

	t.Run("omits roles when there are none", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   testSPCreate,
			ExpectedReqBody:   createServicePrincipalRequest{Name: testSPName},
			ResponseCode:      http.StatusCreated,
			ResponseBody:      spMember(testSPName, testSPLogin),
		})
		sp, err := c.CreateServicePrincipal(t.Context(), testSPOrg, testSPName, nil)
		require.NoError(t, err)
		assert.Equal(t, testSPLogin, sp.Login)
	})

	t.Run("surfaces the service's error", func(t *testing.T) {
		msg := "without custom roles, a service principal holds at most one of the Reader, Maintainer or Admin roles"
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   testSPCreate,
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
	members := spMembers

	t.Run("reads the name from the member list and the roles", func(t *testing.T) {
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			switch {
			case r.Method == http.MethodGet && r.URL.Path == "/api/orgs/an-org/members":
				assert.Equal(t, "service-principal", r.URL.Query().Get("type"))
				return http.StatusOK, members("service-principal:other", testSPLogin)
			case r.Method == http.MethodGet && r.URL.Path == testSPPath+"/roles":
				return http.StatusOK, spRolesResponse(testSPRoleA, testSPRoleB)
			}
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			return http.StatusInternalServerError, nil
		})
		sp, err := c.GetServicePrincipal(t.Context(), testSPOrg, testSPLogin)
		require.NoError(t, err)
		assert.Equal(t, &ServicePrincipal{
			Login:   testSPLogin,
			Name:    "name of " + testSPLogin,
			RoleIDs: []string{testSPRoleA, testSPRoleB},
		}, sp)
	})

	t.Run("follows the member list's continuation token", func(t *testing.T) {
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			if r.URL.Path == testSPPath+"/roles" {
				return http.StatusOK, spRolesResponse()
			}
			if r.URL.Query().Get("continuationToken") == "" {
				page := members("service-principal:other")
				page.ContinuationToken = new("next")
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
				return http.StatusNotFound, ErrorResponse{StatusCode: http.StatusNotFound, Message: notFoundError}
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
		ExpectedReqBody:   renameServicePrincipalRequest{Name: "new-name"},
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
			ExpectedReqBody:   servicePrincipalRolesRequest{RoleIDs: []string{testSPRoleA}},
			ResponseCode:      http.StatusOK,
			ResponseBody:      spRolesResponse(testSPRoleA),
		})
		ids, err := c.UpdateServicePrincipalRoles(t.Context(), testSPOrg, testSPLogin, []string{testSPRoleA})
		require.NoError(t, err)
		assert.Equal(t, []string{testSPRoleA}, ids)
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
