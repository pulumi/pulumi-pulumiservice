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
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"
)

// ServicePrincipalClient manages service principals and their access tokens.
// A service principal is a member of the organization, so it is listed,
// renamed and deleted through the organization member routes, by its login.
type ServicePrincipalClient interface {
	CreateServicePrincipal(ctx context.Context, orgName, name string, roleIDs []string) (*ServicePrincipal, error)
	// GetServicePrincipal returns (nil, nil) when the organization has no service
	// principal with the login.
	GetServicePrincipal(ctx context.Context, orgName, login string) (*ServicePrincipal, error)
	RenameServicePrincipal(ctx context.Context, orgName, login, name string) error
	UpdateServicePrincipalRoles(ctx context.Context, orgName, login string, roleIDs []string) ([]string, error)
	DeleteServicePrincipal(ctx context.Context, orgName, login string) error

	CreateServicePrincipalToken(
		ctx context.Context, orgName, login string, req CreateServicePrincipalTokenRequest,
	) (*AccessToken, error)
	// GetServicePrincipalToken returns (nil, nil) when the token or the service
	// principal doesn't exist.
	GetServicePrincipalToken(ctx context.Context, orgName, login, tokenID string) (*AccessToken, error)
	DeleteServicePrincipalToken(ctx context.Context, orgName, login, tokenID string) error
}

// ServicePrincipal is a service principal of an organization.
type ServicePrincipal struct {
	// Login identifies the service principal in API paths. Unlike Name, it
	// doesn't change when the service principal is renamed.
	Login   string
	Name    string
	RoleIDs []string
}

type createServicePrincipalRequest struct {
	Name    string   `json:"name"`
	RoleIDs []string `json:"roleIds,omitempty"`
}

type servicePrincipalMember struct {
	User User `json:"user"`
}

type renameServicePrincipalRequest struct {
	Name string `json:"name"`
}

type servicePrincipalRolesRequest struct {
	RoleIDs []string `json:"roleIds"`
}

type servicePrincipalRole struct {
	ID string `json:"id"`
}

type servicePrincipalRolesResponse struct {
	Roles []servicePrincipalRole `json:"roles"`
}

// CreateServicePrincipalTokenRequest is the body of a request to create a
// service principal access token.
type CreateServicePrincipalTokenRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Expires     *time.Time `json:"expires,omitempty"`
}

type listTokensPage struct {
	listTokenResponse
	ContinuationToken *string `json:"continuationToken,omitempty"`
}

func servicePrincipalPath(orgName, login string, rest ...string) string {
	return path.Join(append([]string{"orgs", orgName, "service-principals", login}, rest...)...)
}

var (
	errEmptyServicePrincipalOrg   = errors.New("organization name must not be empty")
	errEmptyServicePrincipalLogin = errors.New("service principal login must not be empty")
	errEmptyServicePrincipalName  = errors.New("service principal name must not be empty")
	errEmptyServicePrincipalToken = errors.New("token ID must not be empty")
)

func validateServicePrincipalArgs(orgName, login string) error {
	if orgName == "" {
		return errEmptyServicePrincipalOrg
	}
	if login == "" {
		return errEmptyServicePrincipalLogin
	}
	return nil
}

func (c *Client) CreateServicePrincipal(
	ctx context.Context,
	orgName, name string,
	roleIDs []string,
) (*ServicePrincipal, error) {
	if orgName == "" {
		return nil, errEmptyServicePrincipalOrg
	}
	if name == "" {
		return nil, errEmptyServicePrincipalName
	}
	var res servicePrincipalMember
	apiPath := path.Join("orgs", orgName, "service-principals", "create")
	req := createServicePrincipalRequest{Name: name, RoleIDs: roleIDs}
	if _, err := c.do(ctx, http.MethodPost, apiPath, req, &res); err != nil {
		return nil, fmt.Errorf("failed to create service principal %q: %w", name, err)
	}
	if res.User.GithubLogin == "" {
		return nil, fmt.Errorf("no login returned for service principal %q", name)
	}
	return &ServicePrincipal{
		Login:   res.User.GithubLogin,
		Name:    res.User.Name,
		RoleIDs: roleIDs,
	}, nil
}

func (c *Client) GetServicePrincipal(ctx context.Context, orgName, login string) (*ServicePrincipal, error) {
	if err := validateServicePrincipalArgs(orgName, login); err != nil {
		return nil, err
	}
	members, err := c.listOrgMembersByType(ctx, orgName, "service-principal")
	if err != nil {
		return nil, fmt.Errorf("failed to list service principals: %w", err)
	}
	var found *Member
	for i := range members {
		if members[i].User.GithubLogin == login {
			found = &members[i]
			break
		}
	}
	if found == nil {
		return nil, nil
	}

	var roles servicePrincipalRolesResponse
	if _, err := c.do(ctx, http.MethodGet, servicePrincipalPath(orgName, login, "roles"), nil, &roles); err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list roles of service principal %q: %w", login, err)
	}
	roleIDs := make([]string, 0, len(roles.Roles))
	for _, r := range roles.Roles {
		roleIDs = append(roleIDs, r.ID)
	}
	return &ServicePrincipal{
		Login:   login,
		Name:    found.User.Name,
		RoleIDs: roleIDs,
	}, nil
}

func (c *Client) RenameServicePrincipal(ctx context.Context, orgName, login, name string) error {
	if err := validateServicePrincipalArgs(orgName, login); err != nil {
		return err
	}
	if name == "" {
		return errEmptyServicePrincipalName
	}
	apiPath := path.Join("orgs", orgName, "members", login)
	if _, err := c.do(ctx, http.MethodPatch, apiPath, renameServicePrincipalRequest{Name: name}, nil); err != nil {
		return fmt.Errorf("failed to rename service principal %q: %w", login, err)
	}
	return nil
}

func (c *Client) UpdateServicePrincipalRoles(
	ctx context.Context,
	orgName, login string,
	roleIDs []string,
) ([]string, error) {
	if err := validateServicePrincipalArgs(orgName, login); err != nil {
		return nil, err
	}
	// An empty list removes every role, so send [] rather than null.
	if roleIDs == nil {
		roleIDs = []string{}
	}
	var res servicePrincipalRolesResponse
	req := servicePrincipalRolesRequest{RoleIDs: roleIDs}
	if _, err := c.do(ctx, http.MethodPut, servicePrincipalPath(orgName, login, "roles"), req, &res); err != nil {
		return nil, fmt.Errorf("failed to update roles of service principal %q: %w", login, err)
	}
	ids := make([]string, 0, len(res.Roles))
	for _, r := range res.Roles {
		ids = append(ids, r.ID)
	}
	return ids, nil
}

func (c *Client) DeleteServicePrincipal(ctx context.Context, orgName, login string) error {
	if err := validateServicePrincipalArgs(orgName, login); err != nil {
		return err
	}
	apiPath := path.Join("orgs", orgName, "members", login)
	if _, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil); err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("failed to delete service principal %q: %w", login, err)
	}
	return nil
}

func (c *Client) CreateServicePrincipalToken(
	ctx context.Context,
	orgName, login string,
	req CreateServicePrincipalTokenRequest,
) (*AccessToken, error) {
	if err := validateServicePrincipalArgs(orgName, login); err != nil {
		return nil, err
	}
	if req.Name == "" {
		return nil, errors.New("empty name")
	}
	var res createTokenResponse
	if _, err := c.do(ctx, http.MethodPost, servicePrincipalPath(orgName, login, "tokens"), req, &res); err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}
	token := &AccessToken{
		ID:          res.ID,
		Name:        req.Name,
		TokenValue:  res.TokenValue,
		Description: req.Description,
	}
	if req.Expires != nil {
		token.Expires = req.Expires.Unix()
	}
	return token, nil
}

func (c *Client) GetServicePrincipalToken(
	ctx context.Context,
	orgName, login, tokenID string,
) (*AccessToken, error) {
	if err := validateServicePrincipalArgs(orgName, login); err != nil {
		return nil, err
	}
	if tokenID == "" {
		return nil, errEmptyServicePrincipalToken
	}
	apiPath := servicePrincipalPath(orgName, login, "tokens")
	continuation := ""
	for {
		// The list defaults to active tokens only. Ask for expired ones too, so
		// that a token past its expiry still reads as existing.
		q := url.Values{"filter": []string{"all"}, "search": []string{tokenID}}
		if continuation != "" {
			q.Set("continuationToken", continuation)
		}
		var page listTokensPage
		if _, err := c.doWithQuery(ctx, http.MethodGet, apiPath, q, nil, &page); err != nil {
			if GetErrorStatusCode(err) == http.StatusNotFound {
				return nil, nil
			}
			return nil, fmt.Errorf("failed to list access tokens of service principal %q: %w", login, err)
		}
		for _, t := range page.Tokens {
			if t.ID == tokenID {
				return &AccessToken{
					ID:          t.ID,
					Name:        t.Name,
					Description: t.Description,
					Expires:     t.Expires,
				}, nil
			}
		}
		if page.ContinuationToken == nil || *page.ContinuationToken == "" {
			return nil, nil
		}
		continuation = *page.ContinuationToken
	}
}

func (c *Client) DeleteServicePrincipalToken(ctx context.Context, orgName, login, tokenID string) error {
	if err := validateServicePrincipalArgs(orgName, login); err != nil {
		return err
	}
	if tokenID == "" {
		return errEmptyServicePrincipalToken
	}
	apiPath := servicePrincipalPath(orgName, login, "tokens", tokenID)
	if _, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil); err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("failed to delete access token %q: %w", tokenID, err)
	}
	return nil
}
