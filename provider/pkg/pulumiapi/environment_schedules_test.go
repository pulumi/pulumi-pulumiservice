package pulumiapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testEnv = EnvironmentIdentifier{
	OrgName:     "org",
	ProjectName: "project",
	EnvName:     "env",
}
var createEnvironmentRotationScheduleReq = CreateEnvironmentRotationScheduleRequest{
	ScheduleCron: &cron,
	ScheduleOnce: nil,
}
var testEnvResponse = EnvironmentScheduleResponse{
	ID:           testScheduleID,
	ScheduleOnce: nil,
	ScheduleCron: &cron,
	Definition: EnvironmentScheduleDefinition{
		EnvironmentPath: "",
		EnvironmentID:   "test-env-id",
	},
}

func TestCreateEnvironmentRotationSchedule(t *testing.T) {

	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/esc/environments/org/project/env/schedules",
			ExpectedReqBody:   createEnvironmentRotationScheduleReq,
			ResponseCode:      201,
			ResponseBody:      testEnvResponse,
		})
		expectedScheduleID, err := c.CreateEnvironmentRotationSchedule(
			ctx,
			testEnv,
			createEnvironmentRotationScheduleReq,
		)
		assert.NoError(t, err)
		assert.Equal(t, testScheduleID, *expectedScheduleID)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/esc/environments/org/project/env/schedules",
			ExpectedReqBody:   createEnvironmentRotationScheduleReq,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		expectedScheduleID, err := c.CreateEnvironmentRotationSchedule(
			ctx,
			testEnv,
			createEnvironmentRotationScheduleReq,
		)
		assert.Nil(t, expectedScheduleID, "deployment schedule should be nil since error was returned")
		assert.EqualError(
			t,
			err,
			"failed to create environment rotation schedule (scheduleCron=0 * 0 * 0, scheduleOnce=<nil>): "+
				"401 API error: unauthorized",
		)
	})
}

func TestGetEnvironmentSchedule(t *testing.T) {

	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/esc/environments/org/project/env/schedules/" + testScheduleID,
			ResponseCode:      200,
			ResponseBody:      testEnvResponse,
		})
		response, err := c.GetEnvironmentSchedule(ctx, testEnv, testScheduleID)
		assert.NoError(t, err)
		assert.Equal(t, testEnvResponse, *response)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/esc/environments/org/project/env/schedules/" + testScheduleID,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		expectedScheduleID, err := c.GetEnvironmentSchedule(ctx, testEnv, testScheduleID)
		assert.Nil(t, expectedScheduleID, "scheduleId should be nil since error was returned")
		assert.EqualError(
			t,
			err,
			"failed to get environment schedule with scheduleId test-schedule-id : 401 API error: unauthorized",
		)
	})

	t.Run("404", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/esc/environments/org/project/env/schedules/" + testScheduleID,
			ResponseCode:      404,
			ResponseBody: ErrorResponse{
				StatusCode: 404,
				Message:    "not found",
			},
		})
		expectedScheduleID, err := c.GetEnvironmentSchedule(ctx, testEnv, testScheduleID)
		assert.Nil(t, expectedScheduleID, "scheduleId should be nil since it was not found")
		assert.NoError(t, err)
	})
}

func TestUpdateEnvironmentRotationSchedule(t *testing.T) {

	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/esc/environments/org/project/env/schedules/" + testScheduleID,
			ExpectedReqBody:   createEnvironmentRotationScheduleReq,
			ResponseCode:      201,
			ResponseBody:      testResponse,
		})
		expectedScheduleID, err := c.UpdateEnvironmentRotationSchedule(
			ctx,
			testEnv,
			createEnvironmentRotationScheduleReq,
			testScheduleID,
		)
		assert.NoError(t, err)
		assert.Equal(t, testScheduleID, *expectedScheduleID)
	})

	t.Run("error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/esc/environments/org/project/env/schedules/" + testScheduleID,
			ExpectedReqBody:   createEnvironmentRotationScheduleReq,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		expectedScheduleID, err := c.UpdateEnvironmentRotationSchedule(
			ctx,
			testEnv,
			createEnvironmentRotationScheduleReq,
			testScheduleID,
		)
		assert.Nil(t, expectedScheduleID, "scheduleId should be nil since error was returned")
		assert.EqualError(
			t,
			err,
			"failed to update environment schedule test-schedule-id (scheduleCron=0 * 0 * 0, "+
				"scheduleOnce=<nil>): 401 API error: unauthorized",
		)
	})
}

func TestDeleteEnvironmentSchedule(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/esc/environments/org/project/env/schedules/" + testScheduleID,
			ResponseCode:      201,
		})
		err := c.DeleteEnvironmentSchedule(ctx, testEnv, testScheduleID)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/esc/environments/org/project/env/schedules/" + testScheduleID,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		err := c.DeleteEnvironmentSchedule(ctx, testEnv, testScheduleID)
		assert.EqualError(
			t,
			err,
			"failed to delete environment schedule with scheduleId test-schedule-id : 401 API error: unauthorized",
		)
	})
}
