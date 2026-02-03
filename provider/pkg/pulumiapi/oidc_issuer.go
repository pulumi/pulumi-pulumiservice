package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"path"
)

type OidcClient interface {
	RegisterOidcIssuer(
		ctx context.Context,
		organization string,
		request OidcIssuerRegistrationRequest,
	) (*OidcIssuerRegistrationResponse, error)
	UpdateOidcIssuer(
		ctx context.Context,
		organization string,
		issuerId string,
		request OidcIssuerUpdateRequest,
	) (*OidcIssuerRegistrationResponse, error)
	GetOidcIssuer(ctx context.Context, organization string, issuerId string) (*OidcIssuerRegistrationResponse, error)
	DeleteOidcIssuer(ctx context.Context, organization string, issuerId string) error
	GetAuthPolicies(ctx context.Context, organization string, issuerId string) (*AuthPolicy, error)
	UpdateAuthPolicies(
		ctx context.Context,
		organization string,
		policyId string,
		request AuthPolicyUpdateRequest,
	) (*AuthPolicy, error)
}

type OidcIssuerRegistrationRequest struct {
	Name          string   `json:"name"`
	URL           string   `json:"url"`
	Thumbprints   []string `json:"thumbprints,omitempty"`
	MaxExpiration *int64   `json:"maxExpiration,omitempty"`
}

type OidcIssuerUpdateRequest struct {
	Name          *string   `json:"name,omitempty"`
	Thumbprints   *[]string `json:"thumbprints,omitempty"`
	MaxExpiration *int64    `json:"maxExpiration,omitempty"`
}

type OidcIssuerRegistrationResponse struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	URL           string   `json:"url"`
	Issuer        string   `json:"issuer"`
	Thumbprints   []string `json:"thumbprints,omitempty"`
	MaxExpiration *int64   `json:"maxExpiration,omitempty"`
}

type AuthPolicy struct {
	ID         string                  `json:"id"`
	Version    int                     `json:"version"`
	Created    *string                 `json:"created,omitempty"`
	Modified   *string                 `json:"modified,omitempty"`
	Definition []*AuthPolicyDefinition `json:"policies"`
}

type AuthPolicyDefinition struct {
	Decision              string            `json:"decision"`
	TokenType             string            `json:"tokenType"`
	TeamName              *string           `json:"teamName,omitempty"`
	UserLogin             *string           `json:"userLogin,omitempty"`
	RunnerID              *string           `json:"runnerID,omitempty"`
	AuthorizedPermissions []string          `json:"authorizedPermissions"`
	Rules                 map[string]string `json:"rules"`
}

type AuthPolicyUpdateRequest struct {
	Definition []AuthPolicyDefinition `json:"policies"`
}

func (c *Client) RegisterOidcIssuer(
	ctx context.Context,
	organization string,
	request OidcIssuerRegistrationRequest,
) (*OidcIssuerRegistrationResponse, error) {
	apiPath := path.Join("orgs", organization, "oidc", "issuers")
	var response = &OidcIssuerRegistrationResponse{}
	_, err := c.do(ctx, http.MethodPost, apiPath, request, response)
	if err != nil {
		return nil, fmt.Errorf("failed to register oidc issuer '%s': %w", request.Name, err)
	}
	return response, nil
}

func (c *Client) UpdateOidcIssuer(
	ctx context.Context,
	organization string,
	issuerId string,
	request OidcIssuerUpdateRequest,
) (*OidcIssuerRegistrationResponse, error) {
	apiPath := path.Join("orgs", organization, "oidc", "issuers", issuerId)
	var response = &OidcIssuerRegistrationResponse{}
	_, err := c.do(ctx, http.MethodPatch, apiPath, request, response)
	if err != nil {
		return nil, fmt.Errorf("failed to update oidc issuer with id '%s': %w", issuerId, err)
	}
	return response, nil
}

func (c *Client) GetOidcIssuer(
	ctx context.Context,
	organization string,
	issuerId string,
) (*OidcIssuerRegistrationResponse, error) {
	apiPath := path.Join("orgs", organization, "oidc", "issuers", issuerId)
	var response = &OidcIssuerRegistrationResponse{}
	result, err := c.do(ctx, http.MethodGet, apiPath, nil, response)
	if err != nil {
		if result.StatusCode == 404 {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get oidc issuer with id '%s': %w", issuerId, err)
	}
	return response, nil
}

func (c *Client) DeleteOidcIssuer(ctx context.Context, organization string, issuerId string) error {
	apiPath := path.Join("orgs", organization, "oidc", "issuers", issuerId)
	result, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		if result.StatusCode == 404 {
			return nil
		}
		return fmt.Errorf("failed to delete oidc issuer with id '%s': %w", issuerId, err)
	}
	return nil
}

func (c *Client) GetAuthPolicies(ctx context.Context, organization string, issuerId string) (*AuthPolicy, error) {
	apiPath := path.Join("orgs", organization, "auth", "policies", "oidcissuers", issuerId)
	var response = &AuthPolicy{}
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, response)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth policies with issuer id '%s': %w", issuerId, err)
	}
	return response, nil
}

func (c *Client) UpdateAuthPolicies(
	ctx context.Context,
	organization string,
	policyId string,
	request AuthPolicyUpdateRequest,
) (*AuthPolicy, error) {
	apiPath := path.Join("orgs", organization, "auth", "policies", policyId)
	var response = &AuthPolicy{}
	_, err := c.do(ctx, http.MethodPatch, apiPath, request, response)
	if err != nil {
		return nil, fmt.Errorf("failed to update auth policies with policy id '%s': %w", policyId, err)
	}
	return response, nil
}
