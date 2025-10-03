package pulumiapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPackageVersion(t *testing.T) {
	source := "pulumi"
	publisher := "test-org"
	name := "test-package"
	version := "1.0.0"

	t.Run("Happy Path", func(t *testing.T) {
		resp := PackageMetadata{
			Name:          name,
			Publisher:     publisher,
			Source:        source,
			Version:       version,
			Title:         "Test Package",
			Description:   "A test package",
			PackageStatus: "ga",
			ReadmeURL:     "https://example.com/readme",
			SchemaURL:     "https://example.com/schema.json",
			CreatedAt:     "2025-01-01T00:00:00Z",
			Visibility:    "private",
			IsFeatured:    false,
		}
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/preview/registry/packages/pulumi/test-org/test-package/versions/1.0.0",
			ResponseCode:      200,
			ResponseBody:      resp,
		})
		defer cleanup()

		result, err := c.GetPackageVersion(teamCtx, source, publisher, name, version)
		assert.NoError(t, err)
		assert.Equal(t, &resp, result)
	})

	t.Run("Error - Not Found", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/preview/registry/packages/pulumi/test-org/test-package/versions/1.0.0",
			ResponseCode:      404,
			ResponseBody: ErrorResponse{
				StatusCode: 404,
				Message:    "package version not found",
			},
		})
		defer cleanup()

		_, err := c.GetPackageVersion(teamCtx, source, publisher, name, version)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get package version")
	})
}

func TestDeletePackageVersion(t *testing.T) {
	source := "pulumi"
	publisher := "test-org"
	name := "test-package"
	version := "1.0.0"

	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/preview/registry/packages/pulumi/test-org/test-package/versions/1.0.0",
			ResponseCode:      204,
		})
		defer cleanup()

		err := c.DeletePackageVersion(teamCtx, source, publisher, name, version)
		assert.NoError(t, err)
	})

	t.Run("Not Found - Should Not Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/preview/registry/packages/pulumi/test-org/test-package/versions/1.0.0",
			ResponseCode:      404,
		})
		defer cleanup()

		err := c.DeletePackageVersion(teamCtx, source, publisher, name, version)
		assert.NoError(t, err) // 404 should not error
	})
}

func TestStartPackagePublish(t *testing.T) {
	source := "pulumi"
	publisher := "test-org"
	name := "test-package"
	version := "1.0.0"

	t.Run("Happy Path", func(t *testing.T) {
		resp := StartPackagePublishResponse{
			OperationID: "op-12345",
			UploadURLs: PackageUploadURLs{
				Schema:                    "https://example.com/upload/schema",
				Index:                     "https://example.com/upload/index",
				InstallationConfiguration: "https://example.com/upload/config",
			},
		}
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/preview/registry/packages/pulumi/test-org/test-package/versions",
			ExpectedReqBody: PackageVersionRequest{
				Version: version,
			},
			ResponseCode: 202,
			ResponseBody: resp,
		})
		defer cleanup()

		req := PackageVersionRequest{Version: version}
		result, err := c.StartPackagePublish(teamCtx, source, publisher, name, req)
		assert.NoError(t, err)
		assert.Equal(t, &resp, result)
		assert.Equal(t, "op-12345", result.OperationID)
	})

	t.Run("Error - Conflict", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/preview/registry/packages/pulumi/test-org/test-package/versions",
			ResponseCode:      409,
			ResponseBody: ErrorResponse{
				StatusCode: 409,
				Message:    "package version already exists",
			},
		})
		defer cleanup()

		req := PackageVersionRequest{Version: version}
		_, err := c.StartPackagePublish(teamCtx, source, publisher, name, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to start package publish")
	})
}

func TestCompletePackagePublish(t *testing.T) {
	source := "pulumi"
	publisher := "test-org"
	name := "test-package"
	version := "1.0.0"
	opID := "op-12345"

	t.Run("Happy Path", func(t *testing.T) {
		resp := PublishPackageVersionCompleteResponse{}
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/preview/registry/packages/pulumi/test-org/test-package/versions/1.0.0/complete",
			ExpectedReqBody: PublishPackageVersionCompleteRequest{
				OperationID: opID,
			},
			ResponseCode: 201,
			ResponseBody: resp,
		})
		defer cleanup()

		req := PublishPackageVersionCompleteRequest{OperationID: opID}
		result, err := c.CompletePackagePublish(teamCtx, source, publisher, name, version, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Error - Invalid Operation ID", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/preview/registry/packages/pulumi/test-org/test-package/versions/1.0.0/complete",
			ResponseCode:      404,
			ResponseBody: ErrorResponse{
				StatusCode: 404,
				Message:    "publish operation not found",
			},
		})
		defer cleanup()

		req := PublishPackageVersionCompleteRequest{OperationID: "invalid-op-id"}
		_, err := c.CompletePackagePublish(teamCtx, source, publisher, name, version, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to complete package publish")
	})
}
