package resources

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServicePackageVersionResource struct {
	Client *pulumiapi.Client
}

type PulumiServicePackageVersionInput struct {
	Source                     string
	Publisher                  string
	Name                       string
	Version                    string
	SchemaContent              string
	IndexContent               string
	InstallationConfigContent  string
	Title                      *string
	Description                *string
	LogoURL                    *string
	RepoURL                    *string
	Category                   *string
	PackageStatus              string
	Visibility                 string
	ReadmeURL                  *string
	SchemaURL                  *string
	PluginDownloadURL          *string
	CreatedAt                  *string
}

func (i *PulumiServicePackageVersionInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["source"] = resource.NewPropertyValue(i.Source)
	pm["publisher"] = resource.NewPropertyValue(i.Publisher)
	pm["name"] = resource.NewPropertyValue(i.Name)
	pm["version"] = resource.NewPropertyValue(i.Version)
	pm["schemaContent"] = resource.NewPropertyValue(i.SchemaContent)
	pm["indexContent"] = resource.NewPropertyValue(i.IndexContent)
	pm["installationConfigContent"] = resource.NewPropertyValue(i.InstallationConfigContent)
	pm["packageStatus"] = resource.NewPropertyValue(i.PackageStatus)
	pm["visibility"] = resource.NewPropertyValue(i.Visibility)

	if i.Title != nil {
		pm["title"] = resource.NewPropertyValue(*i.Title)
	}
	if i.Description != nil {
		pm["description"] = resource.NewPropertyValue(*i.Description)
	}
	if i.LogoURL != nil {
		pm["logoUrl"] = resource.NewPropertyValue(*i.LogoURL)
	}
	if i.RepoURL != nil {
		pm["repoUrl"] = resource.NewPropertyValue(*i.RepoURL)
	}
	if i.Category != nil {
		pm["category"] = resource.NewPropertyValue(*i.Category)
	}
	if i.ReadmeURL != nil {
		pm["readmeURL"] = resource.NewPropertyValue(*i.ReadmeURL)
	}
	if i.SchemaURL != nil {
		pm["schemaURL"] = resource.NewPropertyValue(*i.SchemaURL)
	}
	if i.PluginDownloadURL != nil {
		pm["pluginDownloadURL"] = resource.NewPropertyValue(*i.PluginDownloadURL)
	}
	if i.CreatedAt != nil {
		pm["createdAt"] = resource.NewPropertyValue(*i.CreatedAt)
	}

	return pm
}

func (s *PulumiServicePackageVersionResource) ToPulumiServicePackageVersionInput(inputMap resource.PropertyMap) (*PulumiServicePackageVersionInput, error) {
	input := PulumiServicePackageVersionInput{}

	input.Source = inputMap["source"].StringValue()
	input.Publisher = inputMap["publisher"].StringValue()
	input.Name = inputMap["name"].StringValue()
	input.Version = inputMap["version"].StringValue()
	input.SchemaContent = inputMap["schemaContent"].StringValue()
	input.IndexContent = inputMap["indexContent"].StringValue()
	input.InstallationConfigContent = inputMap["installationConfigContent"].StringValue()

	// Optional string fields
	if inputMap["title"].HasValue() && inputMap["title"].IsString() {
		value := inputMap["title"].StringValue()
		input.Title = &value
	}
	if inputMap["description"].HasValue() && inputMap["description"].IsString() {
		value := inputMap["description"].StringValue()
		input.Description = &value
	}
	if inputMap["logoUrl"].HasValue() && inputMap["logoUrl"].IsString() {
		value := inputMap["logoUrl"].StringValue()
		input.LogoURL = &value
	}
	if inputMap["repoUrl"].HasValue() && inputMap["repoUrl"].IsString() {
		value := inputMap["repoUrl"].StringValue()
		input.RepoURL = &value
	}
	if inputMap["category"].HasValue() && inputMap["category"].IsString() {
		value := inputMap["category"].StringValue()
		input.Category = &value
	}

	// Fields with defaults
	if inputMap["packageStatus"].HasValue() && inputMap["packageStatus"].IsString() {
		input.PackageStatus = inputMap["packageStatus"].StringValue()
	} else {
		input.PackageStatus = "ga"
	}

	if inputMap["visibility"].HasValue() && inputMap["visibility"].IsString() {
		input.Visibility = inputMap["visibility"].StringValue()
	} else {
		input.Visibility = "private"
	}

	// Output-only fields (may be present during Read/Update)
	if inputMap["readmeURL"].HasValue() && inputMap["readmeURL"].IsString() {
		value := inputMap["readmeURL"].StringValue()
		input.ReadmeURL = &value
	}
	if inputMap["schemaURL"].HasValue() && inputMap["schemaURL"].IsString() {
		value := inputMap["schemaURL"].StringValue()
		input.SchemaURL = &value
	}
	if inputMap["pluginDownloadURL"].HasValue() && inputMap["pluginDownloadURL"].IsString() {
		value := inputMap["pluginDownloadURL"].StringValue()
		input.PluginDownloadURL = &value
	}
	if inputMap["createdAt"].HasValue() && inputMap["createdAt"].IsString() {
		value := inputMap["createdAt"].StringValue()
		input.CreatedAt = &value
	}

	return &input, nil
}

func (s *PulumiServicePackageVersionResource) Name() string {
	return "pulumiservice:index:PackageVersion"
}

func (s *PulumiServicePackageVersionResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOldInputs(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	diffs := olds.Diff(news)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes: pulumirpc.DiffResponse_DIFF_NONE,
		}, nil
	}

	dd := plugin.NewDetailedDiffFromObjectDiff(diffs, false)

	detailedDiffs := map[string]*pulumirpc.PropertyDiff{}
	replaces := []string(nil)
	// All identifying properties cause replacement
	replaceProperties := map[string]bool{
		"source":    true,
		"publisher": true,
		"name":      true,
		"version":   true,
	}
	for k, v := range dd {
		if _, ok := replaceProperties[k]; ok {
			v.Kind = v.Kind.AsReplace()
			replaces = append(replaces, k)
		}
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind),
			InputDiff: v.InputDiff,
		}
	}

	changes := pulumirpc.DiffResponse_DIFF_SOME
	if len(replaces) > 0 {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}

	return &pulumirpc.DiffResponse{
		Changes:             changes,
		Replaces:            replaces,
		DetailedDiff:        detailedDiffs,
		HasDetailedDiff:     true,
		DeleteBeforeReplace: true,
	}, nil
}

func (s *PulumiServicePackageVersionResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	source, publisher, name, version, err := parsePackageVersionID(req.Id)
	if err != nil {
		return nil, err
	}
	err = s.Client.DeletePackageVersion(ctx, *source, *publisher, *name, *version)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (s *PulumiServicePackageVersionResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputMap, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	input, err := s.ToPulumiServicePackageVersionInput(inputMap)
	if err != nil {
		return nil, err
	}

	// Step 1: Start the publish operation
	startReq := pulumiapi.PackageVersionRequest{
		Version: input.Version,
	}
	startResp, err := s.Client.StartPackagePublish(ctx, input.Source, input.Publisher, input.Name, startReq)
	if err != nil {
		return nil, fmt.Errorf("failed to start package publish: %w", err)
	}

	// Step 2: Upload package artifacts to the provided URLs
	if err := uploadContent(ctx, startResp.UploadURLs.Schema, input.SchemaContent); err != nil {
		return nil, fmt.Errorf("failed to upload schema: %w", err)
	}
	if err := uploadContent(ctx, startResp.UploadURLs.Index, input.IndexContent); err != nil {
		return nil, fmt.Errorf("failed to upload index: %w", err)
	}
	if err := uploadContent(ctx, startResp.UploadURLs.InstallationConfiguration, input.InstallationConfigContent); err != nil {
		return nil, fmt.Errorf("failed to upload installation configuration: %w", err)
	}

	// Step 3: Complete the publish operation
	completeReq := pulumiapi.PublishPackageVersionCompleteRequest{
		OperationID: startResp.OperationID,
	}
	_, err = s.Client.CompletePackagePublish(ctx, input.Source, input.Publisher, input.Name, input.Version, completeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to complete package publish: %w", err)
	}

	// Step 4: Read back the created package to get full metadata
	metadata, err := s.Client.GetPackageVersion(ctx, input.Source, input.Publisher, input.Name, input.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get package metadata after creation: %w", err)
	}

	outputProperties, err := plugin.MarshalProperties(
		toPackageProperties(*metadata).ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
			KeepSecrets:  true,
		},
	)
	if err != nil {
		return nil, err
	}

	id := path.Join(input.Source, input.Publisher, input.Name, input.Version)
	return &pulumirpc.CreateResponse{
		Id:         id,
		Properties: outputProperties,
	}, nil
}

func (s *PulumiServicePackageVersionResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (s *PulumiServicePackageVersionResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	// Package versions are immutable - any change should trigger a replacement
	// This should not be called due to the Diff() implementation marking all key fields as replace-on-change
	return nil, fmt.Errorf("package versions are immutable and cannot be updated; changes require replacement")
}

func (s *PulumiServicePackageVersionResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()
	source, publisher, name, version, err := parsePackageVersionID(req.Id)
	if err != nil {
		return nil, err
	}

	metadata, err := s.Client.GetPackageVersion(ctx, *source, *publisher, *name, *version)
	if err != nil {
		// Check if it's a 404 - package doesn't exist
		if strings.Contains(err.Error(), "404") {
			return &pulumirpc.ReadResponse{}, nil
		}
		return nil, fmt.Errorf("failed to get package version during Read: %w", err)
	}

	properties, err := plugin.MarshalProperties(
		toPackageProperties(*metadata).ToPropertyMap(),
		plugin.MarshalOptions{},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: properties,
		Inputs:     properties,
	}, nil
}

func parsePackageVersionID(id string) (source *string, publisher *string, name *string, version *string, err error) {
	splitID := strings.Split(id, "/")
	if len(splitID) != 4 {
		return nil, nil, nil, nil, fmt.Errorf("invalid package version id: %s (expected format: source/publisher/name/version)", id)
	}
	return &splitID[0], &splitID[1], &splitID[2], &splitID[3], nil
}

func toPackageProperties(metadata pulumiapi.PackageMetadata) *PulumiServicePackageVersionInput {
	input := &PulumiServicePackageVersionInput{
		Source:        metadata.Source,
		Publisher:     metadata.Publisher,
		Name:          metadata.Name,
		Version:       metadata.Version,
		PackageStatus: metadata.PackageStatus,
		Visibility:    metadata.Visibility,
		ReadmeURL:     &metadata.ReadmeURL,
		SchemaURL:     &metadata.SchemaURL,
		CreatedAt:     &metadata.CreatedAt,
	}

	if metadata.Title != "" {
		input.Title = &metadata.Title
	}
	if metadata.Description != "" {
		input.Description = &metadata.Description
	}
	if metadata.LogoURL != "" {
		input.LogoURL = &metadata.LogoURL
	}
	if metadata.RepoURL != "" {
		input.RepoURL = &metadata.RepoURL
	}
	if metadata.Category != "" {
		input.Category = &metadata.Category
	}
	if metadata.PluginDownloadURL != "" {
		input.PluginDownloadURL = &metadata.PluginDownloadURL
	}

	return input
}

// uploadContent uploads content to a pre-signed URL
func uploadContent(ctx context.Context, url string, content string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBufferString(content))
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload content: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
