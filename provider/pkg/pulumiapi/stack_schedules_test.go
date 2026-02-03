package pulumiapi

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testStack = StackIdentifier{
	OrgName:     "org",
	ProjectName: "project",
	StackName:   "stack",
}
var testScheduleID = "test-schedule-id"
var cron = "0 * 0 * 0"
var timestamp = time.Now()
var createDeploymentScheduleReq = CreateDeploymentScheduleRequest{
	ScheduleCron: &cron,
	ScheduleOnce: nil,
	Request: CreateDeploymentRequest{
		PulumiOperation: "update",
	},
}
var createDriftScheduleReq = CreateDriftScheduleRequest{
	ScheduleCron:  cron,
	AutoRemediate: true,
}
var createTtlScheduleReq = CreateTtlScheduleRequest{
	Timestamp:          timestamp,
	DeleteAfterDestroy: true,
}
var testResponse = StackScheduleResponse{
	ID:           testScheduleID,
	ScheduleOnce: nil,
	ScheduleCron: &cron,
	Definition: StackScheduleDefinition{
		Request: CreateDeploymentRequest{
			PulumiOperation: "update",
		},
	},
}

func TestCreateDeploymentSchedule(t *testing.T) {

	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules",
			ExpectedReqBody:   createDeploymentScheduleReq,
			ResponseCode:      201,
			ResponseBody:      testResponse,
		})
		expectedScheduleID, err := c.CreateDeploymentSchedule(ctx, testStack, createDeploymentScheduleReq)
		assert.NoError(t, err)
		assert.Equal(t, testScheduleID, *expectedScheduleID)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules",
			ExpectedReqBody:   createDeploymentScheduleReq,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		expectedScheduleID, err := c.CreateDeploymentSchedule(ctx, testStack, createDeploymentScheduleReq)
		assert.Nil(t, expectedScheduleID, "deployment schedule should be nil since error was returned")
		assert.EqualError(
			t,
			err,
			"failed to create deployment schedule (scheduleCron=0 * 0 * 0, scheduleOnce=<nil>, pulumiOperation=update): 401 API error: unauthorized",
		)
	})
}

func TestGetDeploymentSchedule(t *testing.T) {

	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules/" + testScheduleID,
			ResponseCode:      200,
			ResponseBody:      testResponse,
		})
		response, err := c.GetStackSchedule(ctx, testStack, testScheduleID)
		assert.NoError(t, err)
		assert.Equal(t, testResponse, *response)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules/" + testScheduleID,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		expectedScheduleID, err := c.GetStackSchedule(ctx, testStack, testScheduleID)
		assert.Nil(t, expectedScheduleID, "scheduleId should be nil since error was returned")
		assert.EqualError(
			t,
			err,
			"failed to get stack schedule with scheduleId test-schedule-id : 401 API error: unauthorized",
		)
	})

	t.Run("404", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules/" + testScheduleID,
			ResponseCode:      404,
			ResponseBody: ErrorResponse{
				StatusCode: 404,
				Message:    "not found",
			},
		})
		expectedScheduleID, err := c.GetStackSchedule(ctx, testStack, testScheduleID)
		assert.Nil(t, expectedScheduleID, "scheduleId should be nil since it was not found")
		assert.NoError(t, err)
	})
}

func TestUpdateDeploymentSchedule(t *testing.T) {

	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules/" + testScheduleID,
			ExpectedReqBody:   createDeploymentScheduleReq,
			ResponseCode:      201,
			ResponseBody:      testResponse,
		})
		expectedScheduleID, err := c.UpdateDeploymentSchedule(
			ctx,
			testStack,
			createDeploymentScheduleReq,
			testScheduleID,
		)
		assert.NoError(t, err)
		assert.Equal(t, testScheduleID, *expectedScheduleID)
	})

	t.Run("error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules/" + testScheduleID,
			ExpectedReqBody:   createDeploymentScheduleReq,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		expectedScheduleID, err := c.UpdateDeploymentSchedule(
			ctx,
			testStack,
			createDeploymentScheduleReq,
			testScheduleID,
		)
		assert.Nil(t, expectedScheduleID, "scheduleId should be nil since error was returned")
		assert.EqualError(
			t,
			err,
			"failed to update deployment schedule test-schedule-id (scheduleCron=0 * 0 * 0, scheduleOnce=<nil>, pulumiOperation=update): 401 API error: unauthorized",
		)
	})
}

func TestDeleteSchedule(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules/" + testScheduleID,
			ResponseCode:      201,
		})
		err := c.DeleteStackSchedule(ctx, testStack, testScheduleID)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/schedules/" + testScheduleID,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		err := c.DeleteStackSchedule(ctx, testStack, testScheduleID)
		assert.EqualError(
			t,
			err,
			"failed to delete stack schedule with scheduleId test-schedule-id : 401 API error: unauthorized",
		)
	})
}

func TestCreateDriftSchedule(t *testing.T) {

	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/drift/schedules",
			ExpectedReqBody:   createDriftScheduleReq,
			ResponseCode:      201,
			ResponseBody:      testResponse,
		})
		expectedScheduleID, err := c.CreateDriftSchedule(ctx, testStack, createDriftScheduleReq)
		assert.NoError(t, err)
		assert.Equal(t, testScheduleID, *expectedScheduleID)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/drift/schedules",
			ExpectedReqBody:   createDriftScheduleReq,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		expectedScheduleID, err := c.CreateDriftSchedule(ctx, testStack, createDriftScheduleReq)
		assert.Nil(t, expectedScheduleID, "drift schedule should be nil since error was returned")
		assert.EqualError(
			t,
			err,
			"failed to create drift schedule (scheduleCron=0 * 0 * 0, autoRemediate=true): 401 API error: unauthorized",
		)
	})
}

func TestUpdateDriftSchedule(t *testing.T) {

	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/drift/schedules/" + testScheduleID,
			ExpectedReqBody:   createDriftScheduleReq,
			ResponseCode:      201,
			ResponseBody:      testResponse,
		})
		expectedScheduleID, err := c.UpdateDriftSchedule(ctx, testStack, createDriftScheduleReq, testScheduleID)
		assert.NoError(t, err)
		assert.Equal(t, testScheduleID, *expectedScheduleID)
	})

	t.Run("error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/drift/schedules/" + testScheduleID,
			ExpectedReqBody:   createDriftScheduleReq,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		expectedScheduleID, err := c.UpdateDriftSchedule(ctx, testStack, createDriftScheduleReq, testScheduleID)
		assert.Nil(t, expectedScheduleID, "scheduleId should be nil since error was returned")
		assert.EqualError(
			t,
			err,
			"failed to update drift schedule test-schedule-id (scheduleCron=0 * 0 * 0, autoRemediate=true): 401 API error: unauthorized",
		)
	})
}

func TestCreateTtlSchedule(t *testing.T) {

	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/ttl/schedules",
			ExpectedReqBody:   createTtlScheduleReq,
			ResponseCode:      201,
			ResponseBody:      testResponse,
		})
		expectedScheduleID, err := c.CreateTtlSchedule(ctx, testStack, createTtlScheduleReq)
		assert.NoError(t, err)
		assert.Equal(t, testScheduleID, *expectedScheduleID)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/ttl/schedules",
			ExpectedReqBody:   createTtlScheduleReq,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		expectedScheduleID, err := c.CreateTtlSchedule(ctx, testStack, createTtlScheduleReq)
		assert.Nil(t, expectedScheduleID, "ttl schedule should be nil since error was returned")
		assert.EqualError(
			t,
			err,
			"failed to create ttl schedule (timestamp="+timestamp.String()+", deleteAfterDestroy=true): 401 API error: unauthorized",
		)
	})
}

func TestUpdateTtlSchedule(t *testing.T) {

	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/ttl/schedules/" + testScheduleID,
			ExpectedReqBody:   createTtlScheduleReq,
			ResponseCode:      201,
			ResponseBody:      testResponse,
		})
		expectedScheduleID, err := c.UpdateTtlSchedule(ctx, testStack, createTtlScheduleReq, testScheduleID)
		assert.NoError(t, err)
		assert.Equal(t, testScheduleID, *expectedScheduleID)
	})

	t.Run("error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/stacks/org/project/stack/deployments/ttl/schedules/" + testScheduleID,
			ExpectedReqBody:   createTtlScheduleReq,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		expectedScheduleID, err := c.UpdateTtlSchedule(ctx, testStack, createTtlScheduleReq, testScheduleID)
		assert.Nil(t, expectedScheduleID, "scheduleId should be nil since error was returned")
		assert.EqualError(
			t,
			err,
			"failed to update ttl schedule test-schedule-id (timestamp="+timestamp.String()+", deleteAfterDestroy=true): 401 API error: unauthorized",
		)
	})
}
