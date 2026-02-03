package pulumiapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testInsightsOrgName     = "test-org"
	testInsightsAccountName = "test-account"
)

func TestCreateInsightsAccount(t *testing.T) {
	orgName := testInsightsOrgName
	accountName := testInsightsAccountName

	t.Run("Happy Path", func(t *testing.T) {
		reqBody := CreateInsightsAccountRequest{
			Provider:     "aws",
			Environment:  "test-env",
			ScanSchedule: "daily",
			ProviderConfig: map[string]interface{}{
				"regions": []interface{}{"us-west-2"},
			},
		}

		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s", orgName, accountName),
			ExpectedReqBody:   reqBody,
			ResponseCode:      201,
		})

		err := c.CreateInsightsAccount(t.Context(), orgName, accountName, reqBody)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		reqBody := CreateInsightsAccountRequest{
			Provider:    "aws",
			Environment: "test-env",
		}

		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s", orgName, accountName),
			ExpectedReqBody:   reqBody,
			ResponseCode:      400,
			ResponseBody: ErrorResponse{
				StatusCode: 400,
				Message:    "invalid environment reference",
			},
		})

		err := c.CreateInsightsAccount(t.Context(), orgName, accountName, reqBody)
		assert.EqualError(t, err, `failed to create insights account: 400 API error: invalid environment reference`)
	})
}

func TestListInsightsAccounts(t *testing.T) {
	orgName := testInsightsOrgName

	t.Run("Empty OrgName", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{})

		accounts, err := c.ListInsightsAccounts(t.Context(), "")
		assert.Nil(t, accounts)
		assert.EqualError(t, err, "empty orgName")
	})

	t.Run("Happy Path", func(t *testing.T) {
		resp := ListInsightsAccountsResponse{
			Accounts: []InsightsAccount{
				{
					ID:                   "account-id-123",
					Name:                 "test-account-1",
					Provider:             "aws",
					ProviderEnvRef:       "test-env",
					ScheduledScanEnabled: true,
					ProviderConfig: map[string]interface{}{
						"regions": []interface{}{"us-west-2", "us-east-1"},
					},
				},
				{
					ID:                   "account-id-456",
					Name:                 "test-account-2",
					Provider:             "azure",
					ProviderEnvRef:       "test-env-2",
					ScheduledScanEnabled: false,
				},
			},
		}

		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts", orgName),
			ResponseCode:      200,
			ResponseBody:      resp,
		})

		accounts, err := c.ListInsightsAccounts(t.Context(), orgName)
		require.NoError(t, err)
		assert.Equal(t, 2, len(accounts))
		assert.Equal(t, accounts, resp.Accounts)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts", orgName),
			ResponseCode:      500,
			ResponseBody: ErrorResponse{
				StatusCode: 500,
				Message:    "internal server error",
			},
		})

		accounts, err := c.ListInsightsAccounts(t.Context(), orgName)
		assert.Nil(t, accounts)
		assert.EqualError(t, err, `failed to list insights accounts: 500 API error: internal server error`)
	})
}

func TestGetInsightsAccount(t *testing.T) {
	orgName := testInsightsOrgName
	accountName := testInsightsAccountName

	t.Run("Happy Path", func(t *testing.T) {
		resp := InsightsAccount{
			ID:                   "account-id-123",
			Name:                 accountName,
			Provider:             "aws",
			ProviderEnvRef:       "test-env",
			ScheduledScanEnabled: true,
			ProviderConfig: map[string]interface{}{
				"regions": []interface{}{"us-west-2", "us-east-1"},
			},
		}

		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s", orgName, accountName),
			ResponseCode:      200,
			ResponseBody:      resp,
		})

		account, err := c.GetInsightsAccount(t.Context(), orgName, accountName)
		assert.NoError(t, err)
		assert.Equal(t, &resp, account)
	})

	t.Run("Not Found", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s", orgName, accountName),
			ResponseCode:      404,
			ResponseBody: ErrorResponse{
				StatusCode: 404,
				Message:    "insights account not found",
			},
		})

		account, err := c.GetInsightsAccount(t.Context(), orgName, accountName)
		assert.Nil(t, account)
		assert.NoError(t, err, "404 should return nil, nil")
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s", orgName, accountName),
			ResponseCode:      500,
			ResponseBody: ErrorResponse{
				StatusCode: 500,
				Message:    "internal server error",
			},
		})

		_, err := c.GetInsightsAccount(t.Context(), orgName, accountName)
		assert.EqualError(t, err, `failed to get insights account: 500 API error: internal server error`)
	})
}

func TestUpdateInsightsAccount(t *testing.T) {
	orgName := testInsightsOrgName
	accountName := testInsightsAccountName

	t.Run("Happy Path", func(t *testing.T) {
		reqBody := UpdateInsightsAccountRequest{
			Environment:  "updated-env",
			ScanSchedule: "none",
			ProviderConfig: map[string]interface{}{
				"regions": []interface{}{"us-west-1"},
			},
		}

		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s", orgName, accountName),
			ExpectedReqBody:   reqBody,
			ResponseCode:      204,
		})

		err := c.UpdateInsightsAccount(t.Context(), orgName, accountName, reqBody)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		reqBody := UpdateInsightsAccountRequest{
			Environment: "invalid-env",
		}

		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s", orgName, accountName),
			ExpectedReqBody:   reqBody,
			ResponseCode:      400,
			ResponseBody: ErrorResponse{
				StatusCode: 400,
				Message:    "environment not found",
			},
		})

		err := c.UpdateInsightsAccount(t.Context(), orgName, accountName, reqBody)
		assert.EqualError(t, err, `failed to update insights account: 400 API error: environment not found`)
	})
}

func TestDeleteInsightsAccount(t *testing.T) {
	orgName := testInsightsOrgName
	accountName := testInsightsAccountName

	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s", orgName, accountName),
			ResponseCode:      204,
		})

		err := c.DeleteInsightsAccount(t.Context(), orgName, accountName)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s", orgName, accountName),
			ResponseCode:      404,
			ResponseBody: ErrorResponse{
				StatusCode: 404,
				Message:    "insights account not found",
			},
		})

		err := c.DeleteInsightsAccount(t.Context(), orgName, accountName)
		assert.EqualError(
			t,
			err,
			fmt.Sprintf(`failed to delete insights account "%s": 404 API error: insights account not found`,
				testInsightsAccountName),
		)
	})
}

func TestTriggerScan(t *testing.T) {
	orgName := testInsightsOrgName
	accountName := testInsightsAccountName

	// Note: TriggerScan calls GetScanStatus first to check if a scan is already running.
	// We test GetScanStatus separately, so these tests assume GetScanStatus returns nil
	// (no scan running). In a real scenario, TriggerScan would handle the case where
	// a scan is already running by returning that scan's details.

	t.Run("Happy Path", func(t *testing.T) {
		// Mock server will receive two requests:
		// 1. GET to check current scan status (returns 404 - no scan)
		// 2. POST to trigger new scan
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			expectedPath := fmt.Sprintf("/api/preview/insights/%s/accounts/%s/scan", orgName, accountName)
			assert.Equal(t, expectedPath, r.URL.Path)

			if callCount == 1 {
				// First call: GET to check status
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(404)
				return
			}

			// Second call: POST to trigger scan
			assert.Equal(t, http.MethodPost, r.Method)
			resp := TriggerScanResponse{
				WorkflowRun: WorkflowRun{
					ID:        "run-123",
					Status:    "running",
					StartedAt: "2025-11-18T10:00:00Z",
				},
			}
			w.WriteHeader(200)
			resBytes, _ := json.Marshal(resp)
			_, _ = w.Write(resBytes)
		}))
		defer server.Close()

		httpClient := http.Client{}
		c, err := NewClient(&httpClient, "token", server.URL)
		require.NoError(t, err)

		result, err := c.TriggerScan(t.Context(), orgName, accountName)
		assert.NoError(t, err)
		assert.Equal(t, "run-123", result.ID)
		assert.Equal(t, "running", result.Status)
	})

	t.Run("HTTP 204 No Content - Scan Queued", func(t *testing.T) {
		// Test the case where API returns 204 No Content
		// This means scan was queued but no workflow run details are available yet
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			expectedPath := fmt.Sprintf("/api/preview/insights/%s/accounts/%s/scan", orgName, accountName)
			assert.Equal(t, expectedPath, r.URL.Path)

			if callCount == 1 {
				// First call: GET to check status
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(404)
				return
			}

			// Second call: POST returns 204 No Content
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(204) // No body
		}))
		defer server.Close()

		httpClient := http.Client{}
		c, err := NewClient(&httpClient, "token", server.URL)
		require.NoError(t, err)

		result, err := c.TriggerScan(t.Context(), orgName, accountName)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "queued", result.Status)
	})

	t.Run("Error", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			callCount++
			if callCount == 1 {
				// First call: GET to check status
				w.WriteHeader(404)
				return
			}

			// Second call: POST triggers error
			w.WriteHeader(400)
			errResp := ErrorResponse{
				StatusCode: 400,
				Message:    "scan already in progress",
			}
			resBytes, _ := json.Marshal(errResp)
			_, _ = w.Write(resBytes)
		}))
		defer server.Close()

		httpClient := http.Client{}
		c, err := NewClient(&httpClient, "token", server.URL)
		require.NoError(t, err)

		_, err = c.TriggerScan(t.Context(), orgName, accountName)
		assert.EqualError(
			t,
			err,
			fmt.Sprintf(`failed to trigger scan for insights account "%s": 400 API error: scan already in progress`,
				testInsightsAccountName),
		)
	})
}

func TestGetScanStatus(t *testing.T) {
	orgName := testInsightsOrgName
	accountName := testInsightsAccountName

	t.Run("Happy Path", func(t *testing.T) {
		resp := ScanStatusResponse{
			WorkflowRun: WorkflowRun{
				ID:         "run-456",
				Status:     "succeeded",
				StartedAt:  "2025-11-18T09:00:00Z",
				FinishedAt: "2025-11-18T09:15:00Z",
			},
			NextScan:      "2025-11-19T02:00:00Z",
			ResourceCount: 142,
		}

		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s/scan", orgName, accountName),
			ResponseCode:      200,
			ResponseBody:      resp,
		})

		result, err := c.GetScanStatus(t.Context(), orgName, accountName)
		assert.NoError(t, err)
		assert.Equal(t, &resp, result)
	})

	t.Run("Not Found", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s/scan", orgName, accountName),
			ResponseCode:      404,
			ResponseBody: ErrorResponse{
				StatusCode: 404,
				Message:    "no scan found",
			},
		})

		result, err := c.GetScanStatus(t.Context(), orgName, accountName)
		assert.Nil(t, result)
		assert.NoError(t, err, "404 should return nil, nil")
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s/scan", orgName, accountName),
			ResponseCode:      500,
			ResponseBody: ErrorResponse{
				StatusCode: 500,
				Message:    "internal server error",
			},
		})

		_, err := c.GetScanStatus(t.Context(), orgName, accountName)
		assert.EqualError(
			t,
			err,
			fmt.Sprintf(`failed to get scan status for insights account "%s": 500 API error: internal server error`,
				testInsightsAccountName),
		)
	})
}

func TestGetInsightsAccountTags(t *testing.T) {
	orgName := testInsightsOrgName
	accountName := testInsightsAccountName

	t.Run("Empty OrgName", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/preview/insights//accounts/test-account/tags",
			ResponseCode:      200,
		})

		tags, err := c.GetInsightsAccountTags(t.Context(), "", accountName)
		assert.Nil(t, tags)
		assert.EqualError(t, err, "empty orgName")
	})

	t.Run("Empty AccountName", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/preview/insights/test-org/accounts//tags",
			ResponseCode:      200,
		})

		tags, err := c.GetInsightsAccountTags(t.Context(), orgName, "")
		assert.Nil(t, tags)
		assert.EqualError(t, err, "empty accountName")
	})

	t.Run("Happy Path", func(t *testing.T) {
		resp := GetInsightsAccountTagsResponse{
			Tags: map[string]*InsightsAccountTag{
				"environment": {
					Name:     "environment",
					Value:    "production",
					Created:  "2025-01-01T00:00:00Z",
					Modified: "2025-01-02T00:00:00Z",
				},
				"team": {
					Name:     "team",
					Value:    "platform",
					Created:  "2025-01-01T00:00:00Z",
					Modified: "2025-01-01T00:00:00Z",
				},
			},
		}

		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s/tags", orgName, accountName),
			ResponseCode:      200,
			ResponseBody:      resp,
		})

		tags, err := c.GetInsightsAccountTags(t.Context(), orgName, accountName)
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			"environment": "production",
			"team":        "platform",
		}, tags)
	})

	t.Run("Empty Tags", func(t *testing.T) {
		resp := GetInsightsAccountTagsResponse{
			Tags: map[string]*InsightsAccountTag{},
		}

		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s/tags", orgName, accountName),
			ResponseCode:      200,
			ResponseBody:      resp,
		})

		tags, err := c.GetInsightsAccountTags(t.Context(), orgName, accountName)
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{}, tags)
	})

	t.Run("Nil Tag Value Skipped", func(t *testing.T) {
		resp := GetInsightsAccountTagsResponse{
			Tags: map[string]*InsightsAccountTag{
				"valid-tag": {
					Name:  "valid-tag",
					Value: "value",
				},
				"nil-tag": nil, // This should be skipped
			},
		}

		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s/tags", orgName, accountName),
			ResponseCode:      200,
			ResponseBody:      resp,
		})

		tags, err := c.GetInsightsAccountTags(t.Context(), orgName, accountName)
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			"valid-tag": "value",
		}, tags)
		assert.NotContains(t, tags, "nil-tag")
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s/tags", orgName, accountName),
			ResponseCode:      500,
			ResponseBody: ErrorResponse{
				StatusCode: 500,
				Message:    "internal server error",
			},
		})

		_, err := c.GetInsightsAccountTags(t.Context(), orgName, accountName)
		assert.EqualError(
			t,
			err,
			fmt.Sprintf(`failed to get tags for insights account "%s": 500 API error: internal server error`,
				testInsightsAccountName),
		)
	})
}

func TestSetInsightsAccountTags(t *testing.T) {
	orgName := testInsightsOrgName
	accountName := testInsightsAccountName

	t.Run("Empty OrgName", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPut,
			ExpectedReqPath:   "/api/preview/insights//accounts/test-account/tags",
			ResponseCode:      200,
		})

		err := c.SetInsightsAccountTags(t.Context(), "", accountName, map[string]string{"key": "value"})
		assert.EqualError(t, err, "empty orgName")
	})

	t.Run("Empty AccountName", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPut,
			ExpectedReqPath:   "/api/preview/insights/test-org/accounts//tags",
			ResponseCode:      200,
		})

		err := c.SetInsightsAccountTags(t.Context(), orgName, "", map[string]string{"key": "value"})
		assert.EqualError(t, err, "empty accountName")
	})

	t.Run("Happy Path", func(t *testing.T) {
		reqBody := SetInsightsAccountTagsRequest{
			Tags: map[string]string{
				"environment": "staging",
				"cost-center": "engineering",
			},
		}

		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPut,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s/tags", orgName, accountName),
			ExpectedReqBody:   reqBody,
			ResponseCode:      200,
		})

		err := c.SetInsightsAccountTags(t.Context(), orgName, accountName, reqBody.Tags)
		assert.NoError(t, err)
	})

	t.Run("Empty Tags", func(t *testing.T) {
		reqBody := SetInsightsAccountTagsRequest{
			Tags: map[string]string{},
		}

		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPut,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s/tags", orgName, accountName),
			ExpectedReqBody:   reqBody,
			ResponseCode:      200,
		})

		err := c.SetInsightsAccountTags(t.Context(), orgName, accountName, reqBody.Tags)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		reqBody := SetInsightsAccountTagsRequest{
			Tags: map[string]string{
				"invalid": "tag",
			},
		}

		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPut,
			ExpectedReqPath:   fmt.Sprintf("/api/preview/insights/%s/accounts/%s/tags", orgName, accountName),
			ExpectedReqBody:   reqBody,
			ResponseCode:      400,
			ResponseBody: ErrorResponse{
				StatusCode: 400,
				Message:    "invalid tag name",
			},
		})

		err := c.SetInsightsAccountTags(t.Context(), orgName, accountName, reqBody.Tags)
		assert.EqualError(
			t,
			err,
			fmt.Sprintf(`failed to set tags for insights account "%s": 400 API error: invalid tag name`,
				testInsightsAccountName),
		)
	})
}
