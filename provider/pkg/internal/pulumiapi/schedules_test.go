package pulumiapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testStack = StackName{
	OrgName:     "org",
	ProjectName: "project",
	StackName:   "stack",
}
var testScheduleID = "test-schedule-id"
var cron = "0 * 0 * 0"
var createReq = CreateDeploymentScheduleRequest{
	ScheduleCron: &cron,
	ScheduleOnce: nil,
	Request: CreateDeploymentRequest{
		PulumiOperation: "update",
	},
}
var testResponse = ScheduleResponse{
	ID: testScheduleID,
}

func TestCreateDeploymentSchedule(t *testing.T) {

	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules",
			ExpectedReqBody:   createReq,
			ResponseCode:      201,
			ResponseBody:      testResponse,
		})
		defer cleanup()
		expectedScheduleID, err := c.CreateDeploymentSchedule(ctx, testStack, createReq)
		assert.NoError(t, err)
		assert.Equal(t, testScheduleID, *expectedScheduleID)
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules",
			ExpectedReqBody:   createReq,
			ResponseCode:      401,
			ResponseBody: errorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		expectedScheduleID, err := c.CreateDeploymentSchedule(ctx, testStack, createReq)
		assert.Nil(t, expectedScheduleID, "deployment schedule should be nil since error was returned")
		assert.EqualError(t, err, "failed to create deployment schedule (scheduleCron=0 * 0 * 0, scheduleOnce=<nil>, pulumiOperation=update): 401 API error: unauthorized")
	})
}

func TestGetDeploymentSchedule(t *testing.T) {

	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules/" + testScheduleID,
			ResponseCode:      200,
			ResponseBody:      testResponse,
		})
		defer cleanup()
		expectedScheduleID, err := c.GetSchedule(ctx, testStack, testScheduleID)
		assert.NoError(t, err)
		assert.Equal(t, testScheduleID, *expectedScheduleID)
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules/" + testScheduleID,
			ResponseCode:      401,
			ResponseBody: errorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		expectedScheduleID, err := c.GetSchedule(ctx, testStack, testScheduleID)
		assert.Nil(t, expectedScheduleID, "scheduleID should be nil since error was returned")
		assert.EqualError(t, err, "failed to get schedule with scheduleID test-schedule-id : 401 API error: unauthorized")
	})

	t.Run("404", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules/" + testScheduleID,
			ResponseCode:      404,
			ResponseBody: errorResponse{
				StatusCode: 404,
				Message:    "not found",
			},
		})
		defer cleanup()
		expectedScheduleID, err := c.GetSchedule(ctx, testStack, testScheduleID)
		assert.Nil(t, expectedScheduleID, "scheduleID should be nil since error was returned")
		assert.EqualError(t, err, "failed to get schedule with scheduleID test-schedule-id : 404 API error: not found")
	})
}

func TestUpdateDeploymentSchedule(t *testing.T) {

	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules/" + testScheduleID,
			ExpectedReqBody:   createReq,
			ResponseCode:      201,
			ResponseBody:      testResponse,
		})
		defer cleanup()
		expectedScheduleID, err := c.UpdateDeploymentSchedule(ctx, testStack, createReq, testScheduleID)
		assert.NoError(t, err)
		assert.Equal(t, testScheduleID, *expectedScheduleID)
	})

	t.Run("error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules/" + testScheduleID,
			ExpectedReqBody:   createReq,
			ResponseCode:      401,
			ResponseBody: errorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		expectedScheduleID, err := c.UpdateDeploymentSchedule(ctx, testStack, createReq, testScheduleID)
		assert.Nil(t, expectedScheduleID, "scheduleID should be nil since error was returned")
		assert.EqualError(t, err, "failed to update deployment schedule test-schedule-id (scheduleCron=0 * 0 * 0, scheduleOnce=<nil>, pulumiOperation=update): 401 API error: unauthorized")
	})
}

func TestDeleteSchedule(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules/" + testScheduleID,
			ResponseCode:      201,
		})
		defer cleanup()
		err := c.DeleteSchedule(ctx, testStack, testScheduleID)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules/" + testScheduleID,
			ResponseCode:      401,
			ResponseBody: errorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		err := c.DeleteSchedule(ctx, testStack, testScheduleID)
		assert.EqualError(t, err, "failed to delete schedule with scheduleID test-schedule-id : 401 API error: unauthorized")
	})
}
