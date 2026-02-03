package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"time"
)

type StackScheduleClient interface {
	CreateDeploymentSchedule(
		ctx context.Context,
		stack StackIdentifier,
		req CreateDeploymentScheduleRequest,
	) (*string, error)
	CreateDriftSchedule(ctx context.Context, stack StackIdentifier, req CreateDriftScheduleRequest) (*string, error)
	CreateTtlSchedule(ctx context.Context, stack StackIdentifier, req CreateTtlScheduleRequest) (*string, error)
	GetStackSchedule(ctx context.Context, stack StackIdentifier, scheduleID string) (*StackScheduleResponse, error)
	UpdateDeploymentSchedule(
		ctx context.Context,
		stack StackIdentifier,
		req CreateDeploymentScheduleRequest,
		scheduleID string,
	) (*string, error)
	UpdateDriftSchedule(
		ctx context.Context,
		stack StackIdentifier,
		req CreateDriftScheduleRequest,
		scheduleID string,
	) (*string, error)
	UpdateTtlSchedule(
		ctx context.Context,
		stack StackIdentifier,
		req CreateTtlScheduleRequest,
		scheduleID string,
	) (*string, error)
	DeleteStackSchedule(ctx context.Context, stack StackIdentifier, scheduleID string) error
}

type CreateDeploymentRequest struct {
	PulumiOperation  string                   `json:"operation,omitempty"`
	OperationContext ScheduleOperationContext `json:"operationContext,omitempty"`
}

type ScheduleOperationContext struct {
	Options ScheduleOperationContextOptions `json:"options,omitempty"`
}

type ScheduleOperationContextOptions struct {
	AutoRemediate      bool `json:"remediateIfDriftDetected,omitempty"`
	DeleteAfterDestroy bool `json:"deleteAfterDestroy,omitempty"`
}

type StackScheduleDefinition struct {
	Request CreateDeploymentRequest `json:"request,omitempty"`
}

type CreateDeploymentScheduleRequest struct {
	ScheduleCron *string                 `json:"scheduleCron,omitempty"`
	ScheduleOnce *time.Time              `json:"scheduleOnce,omitempty"`
	Request      CreateDeploymentRequest `json:"request,omitempty"`
}

type CreateDriftScheduleRequest struct {
	ScheduleCron  string `json:"scheduleCron,omitempty"`
	AutoRemediate bool   `json:"autoRemediate,omitempty"`
}

type CreateTtlScheduleRequest struct {
	Timestamp          time.Time `json:"timestamp,omitempty"`
	DeleteAfterDestroy bool      `json:"deleteAfterDestroy,omitempty"`
}

type StackScheduleResponse struct {
	ID           string                  `json:"id,omitempty"`
	ScheduleOnce *string                 `json:"scheduleOnce,omitempty"`
	ScheduleCron *string                 `json:"scheduleCron,omitempty"`
	Definition   StackScheduleDefinition `json:"definition,omitempty"`
}

func (c *Client) CreateDeploymentSchedule(
	ctx context.Context,
	stack StackIdentifier,
	scheduleReq CreateDeploymentScheduleRequest,
) (*string, error) {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "deployments", "schedules")
	var scheduleResponse StackScheduleResponse
	_, err := c.do(ctx, http.MethodPost, apiPath, scheduleReq, &scheduleResponse)
	if err != nil {
		var cronString string
		if scheduleReq.ScheduleCron != nil {
			cronString = *scheduleReq.ScheduleCron
		} else {
			cronString = "<nil>"
		}
		return nil, fmt.Errorf(
			"failed to create deployment schedule (scheduleCron=%s, scheduleOnce=%s, pulumiOperation=%s): %w",
			cronString,
			scheduleReq.ScheduleOnce,
			scheduleReq.Request.PulumiOperation,
			err,
		)
	}
	return &scheduleResponse.ID, nil
}

func (c *Client) CreateDriftSchedule(
	ctx context.Context,
	stack StackIdentifier,
	scheduleReq CreateDriftScheduleRequest,
) (*string, error) {
	apiPath := path.Join(
		"stacks",
		stack.OrgName,
		stack.ProjectName,
		stack.StackName,
		"deployments",
		"drift",
		"schedules",
	)
	var scheduleResponse StackScheduleResponse
	_, err := c.do(ctx, http.MethodPost, apiPath, scheduleReq, &scheduleResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to create drift schedule (scheduleCron=%s, autoRemediate=%t): %w",
			scheduleReq.ScheduleCron, scheduleReq.AutoRemediate, err)
	}
	return &scheduleResponse.ID, nil
}

func (c *Client) CreateTtlSchedule(
	ctx context.Context,
	stack StackIdentifier,
	scheduleReq CreateTtlScheduleRequest,
) (*string, error) {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "deployments", "ttl", "schedules")
	var scheduleResponse StackScheduleResponse
	_, err := c.do(ctx, http.MethodPost, apiPath, scheduleReq, &scheduleResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to create ttl schedule (timestamp=%s, deleteAfterDestroy=%t): %w",
			scheduleReq.Timestamp, scheduleReq.DeleteAfterDestroy, err)
	}
	return &scheduleResponse.ID, nil
}

func (c *Client) GetStackSchedule(
	ctx context.Context,
	stack StackIdentifier,
	scheduleID string,
) (*StackScheduleResponse, error) {
	apiPath := path.Join(
		"stacks",
		stack.OrgName,
		stack.ProjectName,
		stack.StackName,
		"deployments",
		"schedules",
		scheduleID,
	)
	var scheduleResponse StackScheduleResponse
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &scheduleResponse)
	if err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get stack schedule with scheduleId %s : %w", scheduleID, err)
	}
	return &scheduleResponse, nil
}

func (c *Client) UpdateDeploymentSchedule(
	ctx context.Context,
	stack StackIdentifier,
	scheduleReq CreateDeploymentScheduleRequest,
	scheduleID string,
) (*string, error) {
	apiPath := path.Join(
		"stacks",
		stack.OrgName,
		stack.ProjectName,
		stack.StackName,
		"deployments",
		"schedules",
		scheduleID,
	)
	var scheduleResponse StackScheduleResponse
	_, err := c.do(ctx, http.MethodPost, apiPath, scheduleReq, &scheduleResponse)
	if err != nil {
		var cronString string
		if scheduleReq.ScheduleCron != nil {
			cronString = *scheduleReq.ScheduleCron
		} else {
			cronString = "<nil>"
		}
		return nil, fmt.Errorf(
			"failed to update deployment schedule %s (scheduleCron=%s, scheduleOnce=%s, pulumiOperation=%s): %w",
			scheduleID,
			cronString,
			scheduleReq.ScheduleOnce,
			scheduleReq.Request.PulumiOperation,
			err,
		)
	}
	return &scheduleResponse.ID, nil
}

func (c *Client) UpdateDriftSchedule(
	ctx context.Context,
	stack StackIdentifier,
	scheduleReq CreateDriftScheduleRequest,
	scheduleID string,
) (*string, error) {
	apiPath := path.Join(
		"stacks",
		stack.OrgName,
		stack.ProjectName,
		stack.StackName,
		"deployments",
		"drift",
		"schedules",
		scheduleID,
	)
	var scheduleResponse StackScheduleResponse
	_, err := c.do(ctx, http.MethodPost, apiPath, scheduleReq, &scheduleResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to update drift schedule %s (scheduleCron=%s, autoRemediate=%t): %w",
			scheduleID, scheduleReq.ScheduleCron, scheduleReq.AutoRemediate, err)
	}
	return &scheduleResponse.ID, nil
}

func (c *Client) UpdateTtlSchedule(
	ctx context.Context,
	stack StackIdentifier,
	scheduleReq CreateTtlScheduleRequest,
	scheduleID string,
) (*string, error) {
	apiPath := path.Join(
		"stacks",
		stack.OrgName,
		stack.ProjectName,
		stack.StackName,
		"deployments",
		"ttl",
		"schedules",
		scheduleID,
	)
	var scheduleResponse StackScheduleResponse
	_, err := c.do(ctx, http.MethodPost, apiPath, scheduleReq, &scheduleResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to update ttl schedule %s (timestamp=%s, deleteAfterDestroy=%t): %w",
			scheduleID, scheduleReq.Timestamp, scheduleReq.DeleteAfterDestroy, err)
	}
	return &scheduleResponse.ID, nil
}

func (c *Client) DeleteStackSchedule(ctx context.Context, stack StackIdentifier, scheduleID string) error {
	apiPath := path.Join(
		"stacks",
		stack.OrgName,
		stack.ProjectName,
		stack.StackName,
		"deployments",
		"schedules",
		scheduleID,
	)
	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete stack schedule with scheduleId %s : %w", scheduleID, err)
	}
	return nil
}
