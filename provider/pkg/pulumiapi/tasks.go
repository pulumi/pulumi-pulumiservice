// Copyright 2016-2022, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package pulumiapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"time"
)

type TaskClient interface {
	CreateTask(ctx context.Context, orgName string, req CreateTaskRequest) (*Task, error)
	GetTask(ctx context.Context, orgName string, taskID string) (*Task, error)
}

type Task struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Status    string       `json:"status"`
	CreatedAt time.Time    `json:"createdAt"`
	Entities  []TaskEntity `json:"entities"`
}

type TaskEntity struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type EntityDiff struct {
	Add    []TaskEntity `json:"add,omitempty"`
	Remove []TaskEntity `json:"remove,omitempty"`
}

type CreateTaskRequest struct {
	Content    string      `json:"content"`
	EntityDiff *EntityDiff `json:"entity_diff,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
}

func (c *Client) CreateTask(ctx context.Context, orgName string, req CreateTaskRequest) (*Task, error) {
	if len(orgName) == 0 {
		return nil, errors.New("orgName must not be empty")
	}

	if len(req.Content) == 0 {
		return nil, errors.New("content must not be empty")
	}

	if req.Timestamp.IsZero() {
		return nil, errors.New("timestamp must not be empty")
	}

	apiPath := path.Join("preview", "agents", orgName, "tasks")

	var task Task
	_, err := c.do(ctx, http.MethodPost, apiPath, req, &task)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return &task, nil
}

func (c *Client) GetTask(ctx context.Context, orgName string, taskID string) (*Task, error) {
	if len(orgName) == 0 {
		return nil, errors.New("orgName must not be empty")
	}

	if len(taskID) == 0 {
		return nil, errors.New("taskID must not be empty")
	}

	apiPath := path.Join("preview", "agents", orgName, "tasks", taskID)

	var task Task
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &task)
	if err != nil {
		statusCode := GetErrorStatusCode(err)
		if statusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}
