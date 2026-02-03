package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"time"
)

const nilScheduleCronPlaceholder = "<nil>"

type EnvironmentScheduleClient interface {
	CreateEnvironmentRotationSchedule(
		ctx context.Context,
		environment EnvironmentIdentifier,
		req CreateEnvironmentRotationScheduleRequest,
	) (*string, error)
	GetEnvironmentSchedule(
		ctx context.Context,
		environment EnvironmentIdentifier,
		scheduleID string,
	) (*EnvironmentScheduleResponse, error)
	UpdateEnvironmentRotationSchedule(
		ctx context.Context,
		environment EnvironmentIdentifier,
		req CreateEnvironmentRotationScheduleRequest,
		scheduleID string,
	) (*string, error)
	DeleteEnvironmentSchedule(ctx context.Context, environment EnvironmentIdentifier, scheduleID string) error
}

type EnvironmentIdentifier struct {
	OrgName     string `json:"orgName"`
	ProjectName string `json:"projectName"`
	EnvName     string `json:"envName"`
}

type CreateEnvironmentRotationScheduleRequest struct {
	ScheduleCron          *string                                        `json:"scheduleCron,omitempty"`
	ScheduleOnce          *time.Time                                     `json:"scheduleOnce,omitempty"`
	SecretRotationRequest CreateEnvironmentSecretRotationScheduleRequest `json:"secretRotationRequest,omitempty"`
}

type CreateEnvironmentSecretRotationScheduleRequest struct {
	EnvironmentPath *string `json:"environmentPath,omitempty"`
}

type EnvironmentScheduleResponse struct {
	ID           string                        `json:"id,omitempty"`
	ScheduleOnce *string                       `json:"scheduleOnce,omitempty"`
	ScheduleCron *string                       `json:"scheduleCron,omitempty"`
	Definition   EnvironmentScheduleDefinition `json:"definition,omitempty"`
}

type EnvironmentScheduleDefinition struct {
	EnvironmentPath string `json:"environmentPath"`
	EnvironmentID   string `json:"environmentID"`
}

func (c *Client) CreateEnvironmentRotationSchedule(
	ctx context.Context,
	environment EnvironmentIdentifier,
	scheduleReq CreateEnvironmentRotationScheduleRequest,
) (*string, error) {
	apiPath := path.Join(
		"esc",
		"environments",
		environment.OrgName,
		environment.ProjectName,
		environment.EnvName,
		"schedules",
	)
	var scheduleResponse EnvironmentScheduleResponse
	_, err := c.do(ctx, http.MethodPost, apiPath, scheduleReq, &scheduleResponse)
	if err != nil {
		var cronString string
		if scheduleReq.ScheduleCron != nil {
			cronString = *scheduleReq.ScheduleCron
		} else {
			cronString = nilScheduleCronPlaceholder
		}
		return nil, fmt.Errorf("failed to create environment rotation schedule (scheduleCron=%s, scheduleOnce=%s): %w",
			cronString, scheduleReq.ScheduleOnce, err)
	}
	return &scheduleResponse.ID, nil
}

func (c *Client) GetEnvironmentSchedule(
	ctx context.Context,
	environment EnvironmentIdentifier,
	scheduleID string,
) (*EnvironmentScheduleResponse, error) {
	apiPath := path.Join(
		"esc",
		"environments",
		environment.OrgName,
		environment.ProjectName,
		environment.EnvName,
		"schedules",
		scheduleID,
	)
	var scheduleResponse EnvironmentScheduleResponse
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &scheduleResponse)
	if err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get environment schedule with scheduleId %s : %w", scheduleID, err)
	}
	return &scheduleResponse, nil
}

func (c *Client) UpdateEnvironmentRotationSchedule(
	ctx context.Context,
	environment EnvironmentIdentifier,
	scheduleReq CreateEnvironmentRotationScheduleRequest,
	scheduleID string,
) (*string, error) {
	apiPath := path.Join(
		"esc",
		"environments",
		environment.OrgName,
		environment.ProjectName,
		environment.EnvName,
		"schedules",
		scheduleID,
	)
	var scheduleResponse EnvironmentScheduleResponse
	_, err := c.do(ctx, http.MethodPost, apiPath, scheduleReq, &scheduleResponse)
	if err != nil {
		var cronString string
		if scheduleReq.ScheduleCron != nil {
			cronString = *scheduleReq.ScheduleCron
		} else {
			cronString = nilScheduleCronPlaceholder
		}
		return nil, fmt.Errorf("failed to update environment schedule %s (scheduleCron=%s, scheduleOnce=%s): %w",
			scheduleID, cronString, scheduleReq.ScheduleOnce, err)
	}
	return &scheduleResponse.ID, nil
}

func (c *Client) DeleteEnvironmentSchedule(
	ctx context.Context,
	environment EnvironmentIdentifier,
	scheduleID string,
) error {
	apiPath := path.Join(
		"esc",
		"environments",
		environment.OrgName,
		environment.ProjectName,
		environment.EnvName,
		"schedules",
		scheduleID,
	)
	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete environment schedule with scheduleId %s : %w", scheduleID, err)
	}
	return nil
}
