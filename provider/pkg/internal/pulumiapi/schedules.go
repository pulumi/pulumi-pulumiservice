package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"time"
)

type ScheduleClient interface {
	CreateDeploymentSchedule(ctx context.Context, stack StackName, req CreateDeploymentScheduleRequest) (*string, error)
	GetSchedule(ctx context.Context, stack StackName, scheduleID string) (*string, error)
	UpdateDeploymentSchedule(ctx context.Context, stack StackName, req CreateDeploymentScheduleRequest, scheduleID string) (*string, error)
	DeleteSchedule(ctx context.Context, stack StackName, scheduleID string) error
}

type CreateDeploymentRequest struct {
	PulumiOperation string `json:"operation,omitempty"`
}

type ScheduleDefinition struct {
	Request CreateDeploymentRequest `json:"request,omitempty"`
}

type CreateDeploymentScheduleRequest struct {
	ScheduleCron *string                 `json:"scheduleCron,omitempty"`
	ScheduleOnce *time.Time              `json:"scheduleOnce,omitempty"`
	Request      CreateDeploymentRequest `json:"request,omitempty"`
}

type ScheduleResponse struct {
	ID string
}

func (c *Client) CreateDeploymentSchedule(ctx context.Context, stack StackName, scheduleReq CreateDeploymentScheduleRequest) (*string, error) {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "deployments", "schedules")
	var scheduleResponse ScheduleResponse
	_, err := c.do(ctx, http.MethodPost, apiPath, scheduleReq, &scheduleResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment schedule (scheduleCron=%s, scheduleOnce=%s, pulumiOperation=%s): %w",
			*scheduleReq.ScheduleCron, scheduleReq.ScheduleOnce, scheduleReq.Request.PulumiOperation, err)
	}
	return &scheduleResponse.ID, nil
}

func (c *Client) GetSchedule(ctx context.Context, stack StackName, scheduleID string) (*string, error) {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "deployments", "schedules", scheduleID)
	var scheduleResponse ScheduleResponse
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &scheduleResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule with scheduleID %s : %w", scheduleID, err)
	}
	return &scheduleResponse.ID, nil
}

func (c *Client) UpdateDeploymentSchedule(ctx context.Context, stack StackName, scheduleReq CreateDeploymentScheduleRequest, scheduleID string) (*string, error) {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "deployments", "schedules", scheduleID)
	var scheduleResponse ScheduleResponse
	_, err := c.do(ctx, http.MethodPost, apiPath, scheduleReq, &scheduleResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to update deployment schedule %s (scheduleCron=%s, scheduleOnce=%s, pulumiOperation=%s): %w",
			scheduleID, *scheduleReq.ScheduleCron, scheduleReq.ScheduleOnce, scheduleReq.Request.PulumiOperation, err)
	}
	return &scheduleResponse.ID, nil
}

func (c *Client) DeleteSchedule(ctx context.Context, stack StackName, scheduleID string) error {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "deployments", "schedules", scheduleID)
	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete schedule with scheduleID %s : %w", scheduleID, err)
	}
	return nil
}
