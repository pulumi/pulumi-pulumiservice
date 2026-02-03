package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"path"
)

type CreateTemplateSourceRequestDestination struct {
	URL *string `json:"url"`
}

type CreateTemplateSourceRequest struct {
	Name        string                                  `json:"name"`
	SourceURL   string                                  `json:"sourceURL"`
	Destination *CreateTemplateSourceRequestDestination `json:"destination"`
}

type TemplateSourceResponse struct {
	ID          string                                  `json:"id"`
	IsValid     bool                                    `json:"isValid"`
	Name        string                                  `json:"name"`
	SourceURL   string                                  `json:"sourceURL"`
	Destination *CreateTemplateSourceRequestDestination `json:"destination"`
}

type ListResponse struct {
	Sources []TemplateSourceResponse `json:"sources"`
}

func (c *Client) CreateTemplateSource(
	ctx context.Context,
	organizationName string,
	request CreateTemplateSourceRequest,
) (*TemplateSourceResponse, error) {
	apiPath := path.Join("orgs", organizationName, "templates/sources")
	var response TemplateSourceResponse
	_, err := c.do(ctx, http.MethodPost, apiPath, request, &response)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create template source in org %s: %+v due to error: %w",
			organizationName,
			request,
			err,
		)
	}
	return &response, nil
}

func (c *Client) UpdateTemplateSource(
	ctx context.Context,
	organizationName string,
	templateID string,
	request CreateTemplateSourceRequest,
) (*TemplateSourceResponse, error) {
	apiPath := path.Join("orgs", organizationName, "templates/sources", templateID)
	var response TemplateSourceResponse
	_, err := c.do(ctx, http.MethodPatch, apiPath, request, &response)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to update template source in org %s with id %s: %+v due to error: %w",
			organizationName,
			templateID,
			request,
			err,
		)
	}
	return &response, nil
}

func (c *Client) GetTemplateSource(
	ctx context.Context,
	organizationName string,
	templateID string,
) (*TemplateSourceResponse, error) {
	// This sucks, but there's not Get API for Template Sources
	// Thus, using a List and then finding by ID
	// TODO issue to improve this - https://github.com/pulumi/pulumi-service/issues/21637

	apiPath := path.Join("orgs", organizationName, "templates/sources")
	var templateSources ListResponse
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &templateSources)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get template source in org %s with id %s due to error: %w",
			organizationName,
			templateID,
			err,
		)
	}

	for _, source := range templateSources.Sources {
		if source.ID == templateID {
			return &source, nil
		}
	}

	return nil, nil
}

func (c *Client) DeleteTemplateSource(ctx context.Context, organizationName string, templateID string) error {
	apiPath := path.Join("orgs", organizationName, "templates/sources", templateID)
	response, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if response.StatusCode == 404 {
		return nil
	}
	if err != nil {
		return fmt.Errorf(
			"failed to delete template source in org %s with id %s due to error: %w",
			organizationName,
			templateID,
			err,
		)
	}
	return nil
}
