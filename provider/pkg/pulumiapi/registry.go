package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"path"
)

// PackageVersionRequest represents the request to start publishing a package version
type PackageVersionRequest struct {
	Version string `json:"version"`
}

// PackageUploadURLs contains the URLs for uploading package artifacts
type PackageUploadURLs struct {
	Schema                    string `json:"schema"`
	Index                     string `json:"index"`
	InstallationConfiguration string `json:"installationConfiguration"`
}

// StartPackagePublishResponse contains the response from starting a package publish
type StartPackagePublishResponse struct {
	OperationID string            `json:"operationID"`
	UploadURLs  PackageUploadURLs `json:"uploadURLs"`
}

// PublishPackageVersionCompleteRequest represents the request to complete a package publish
type PublishPackageVersionCompleteRequest struct {
	OperationID string `json:"operationID"`
}

// PublishPackageVersionCompleteResponse contains the response from completing a package publish
type PublishPackageVersionCompleteResponse struct {
	// Response fields if any
}

// PackageParameterization contains parameterization information for a package
type PackageParameterization struct {
	BaseProvider string            `json:"baseProvider,omitempty"`
	Parameter    string            `json:"parameter,omitempty"`
	Version      string            `json:"version,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// PackageMetadata represents metadata about a published package version
type PackageMetadata struct {
	Name                string                   `json:"name"`
	Publisher           string                   `json:"publisher"`
	Source              string                   `json:"source"`
	Version             string                   `json:"version"`
	Title               string                   `json:"title,omitempty"`
	Description         string                   `json:"description,omitempty"`
	LogoURL             string                   `json:"logoUrl,omitempty"`
	RepoURL             string                   `json:"repoUrl,omitempty"`
	Category            string                   `json:"category,omitempty"`
	IsFeatured          bool                     `json:"isFeatured"`
	PackageTypes        []string                 `json:"packageTypes,omitempty"`
	PackageStatus       string                   `json:"packageStatus"`
	ReadmeURL           string                   `json:"readmeURL"`
	SchemaURL           string                   `json:"schemaURL"`
	PluginDownloadURL   string                   `json:"pluginDownloadURL,omitempty"`
	CreatedAt           string                   `json:"createdAt"`
	Visibility          string                   `json:"visibility"`
	Parameterization    *PackageParameterization `json:"parameterization,omitempty"`
}

// ListPackagesResponse contains the response from listing packages
type ListPackagesResponse struct {
	Packages          []PackageMetadata `json:"packages"`
	ContinuationToken *string           `json:"continuationToken,omitempty"`
}

// StartPackagePublish initiates a package version publish operation
func (c *Client) StartPackagePublish(ctx context.Context, source, publisher, name string, req PackageVersionRequest) (*StartPackagePublishResponse, error) {
	apiPath := path.Join("preview/registry/packages", source, publisher, name, "versions")
	var response StartPackagePublishResponse
	_, err := c.do(ctx, http.MethodPost, apiPath, req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to start package publish for %s/%s/%s: %w", source, publisher, name, err)
	}
	return &response, nil
}

// CompletePackagePublish completes a package version publish operation
func (c *Client) CompletePackagePublish(ctx context.Context, source, publisher, name, version string, req PublishPackageVersionCompleteRequest) (*PublishPackageVersionCompleteResponse, error) {
	apiPath := path.Join("preview/registry/packages", source, publisher, name, "versions", version, "complete")
	var response PublishPackageVersionCompleteResponse
	_, err := c.do(ctx, http.MethodPost, apiPath, req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to complete package publish for %s/%s/%s@%s: %w", source, publisher, name, version, err)
	}
	return &response, nil
}

// GetPackageVersion retrieves metadata for a specific package version
func (c *Client) GetPackageVersion(ctx context.Context, source, publisher, name, version string) (*PackageMetadata, error) {
	apiPath := path.Join("preview/registry/packages", source, publisher, name, "versions", version)
	var response PackageMetadata
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get package version for %s/%s/%s@%s: %w", source, publisher, name, version, err)
	}
	return &response, nil
}

// DeletePackageVersion deletes a specific package version
func (c *Client) DeletePackageVersion(ctx context.Context, source, publisher, name, version string) error {
	apiPath := path.Join("preview/registry/packages", source, publisher, name, "versions", version)
	response, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if response != nil && response.StatusCode == 404 {
		// Package version already deleted
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to delete package version for %s/%s/%s@%s: %w", source, publisher, name, version, err)
	}
	return nil
}

// ListPackages lists packages in the registry
func (c *Client) ListPackages(ctx context.Context, orgLogin *string, name *string, visibility *string, limit *int, continuationToken *string) (*ListPackagesResponse, error) {
	apiPath := "preview/registry/packages"

	// Build query parameters
	query := make(map[string]string)
	if orgLogin != nil {
		query["orgLogin"] = *orgLogin
	}
	if name != nil {
		query["name"] = *name
	}
	if visibility != nil {
		query["visibility"] = *visibility
	}
	if limit != nil {
		query["limit"] = fmt.Sprintf("%d", *limit)
	}
	if continuationToken != nil {
		query["continuationToken"] = *continuationToken
	}

	var response ListPackagesResponse
	// Note: We'd need to update the client to support query parameters
	// For now, this is a placeholder showing the structure
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}
	return &response, nil
}
