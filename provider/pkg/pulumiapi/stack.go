// Copyright 2016-2026, Pulumi Corporation.
package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// StackConfig matches the Pulumi Cloud Service-Backed Configuration payload
// (apitype.StackConfig). Only `Environment` is used by PSP today; the secrets-
// related fields are accepted by the server but not surfaced as user inputs.
type StackConfig struct {
	Environment     string `json:"environment"`
	SecretsProvider string `json:"secretsProvider,omitempty"`
	EncryptedKey    string `json:"encryptedKey,omitempty"`
	EncryptionSalt  string `json:"encryptionSalt,omitempty"`
}

type CreateStackRequest struct {
	StackName string       `json:"stackName"`
	Config    *StackConfig `json:"config,omitempty"`
}

// CreateStack creates a new stack. When config is non-nil, the server will
// also create the referenced ESC environment (if it doesn't already exist)
// and link it as the stack's service-backed config in a single call. Pass nil
// for the legacy stack-only flow.
func (c *Client) CreateStack(ctx context.Context, stack StackIdentifier, config *StackConfig) error {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName)
	_, err := c.do(ctx, http.MethodPost, apiPath, CreateStackRequest{
		StackName: stack.StackName,
		Config:    config,
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to create stack '%s': %w", stack, err)
	}
	return nil
}

func (c *Client) StackExists(ctx context.Context, stackName StackIdentifier) (bool, error) {
	if stackName.OrgName == "" || stackName.ProjectName == "" || stackName.StackName == "" {
		return false, fmt.Errorf("invalid stack identifier: %v", stackName)
	}
	apiPath := path.Join("stacks", stackName.OrgName, stackName.ProjectName, stackName.StackName)
	var s stack
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &s)
	if err != nil {
		statusCode := GetErrorStatusCode(err)
		if statusCode == http.StatusNotFound {
			return false, nil
		}

		return false, fmt.Errorf("failed to get stack: %w", err)
	}
	return true, nil
}

// DeleteStack deletes a stack. preserveEnvironment controls whether the server
// should keep the linked SBC environment alive on stack delete; PSP sets it to
// true unless we own the env's lifecycle (auto mode).
func (c *Client) DeleteStack(
	ctx context.Context,
	stackName StackIdentifier,
	forceDestroy bool,
	preserveEnvironment bool,
) error {
	apiPath := path.Join(
		"stacks", stackName.OrgName, stackName.ProjectName, stackName.StackName,
	)

	query := url.Values{}
	if forceDestroy {
		query.Set("force", "true")
	}
	if preserveEnvironment {
		query.Set("preserveEnvironment", "true")
	}

	var err error
	if len(query) > 0 {
		_, err = c.doWithQuery(ctx, http.MethodDelete, apiPath, query, nil, nil)
	} else {
		_, err = c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	}
	if err != nil {
		return fmt.Errorf("failed to delete stack: %w", err)
	}

	return nil
}

// GetStackConfig returns the service-backed stack config, or (nil, nil) if no
// such config is set (server replies 404). A non-nil error is reserved for
// real failures.
func (c *Client) GetStackConfig(ctx context.Context, stack StackIdentifier) (*StackConfig, error) {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "config")
	var cfg StackConfig
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &cfg)
	if err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get stack config '%s': %w", stack, err)
	}
	return &cfg, nil
}

// SetStackConfig creates or updates the service-backed stack config link.
func (c *Client) SetStackConfig(ctx context.Context, stack StackIdentifier, config StackConfig) error {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "config")
	_, err := c.do(ctx, http.MethodPut, apiPath, config, nil)
	if err != nil {
		return fmt.Errorf("failed to set stack config '%s': %w", stack, err)
	}
	return nil
}

// DeleteStackConfig removes the service-backed stack config link.
func (c *Client) DeleteStackConfig(ctx context.Context, stack StackIdentifier) error {
	apiPath := path.Join("stacks", stack.OrgName, stack.ProjectName, stack.StackName, "config")
	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("failed to delete stack config '%s': %w", stack, err)
	}
	return nil
}

// FormatEnvRef builds the wire-format ESC environment reference used by the
// stack config payload: `project/env` or `project/env@version` when version is
// set. The server side parses this with environments.ParseRef, which accepts
// both `@` and `:` as the version separator; we always emit `@`.
func FormatEnvRef(project, name, version string) string {
	ref := path.Join(project, name)
	if version != "" {
		ref += "@" + version
	}
	return ref
}

// ParseEnvRef inverts FormatEnvRef. It accepts the legacy two-part form
// (`project/env`) plus an optional `@version` or `:version` suffix on the
// final segment, mirroring esc's CLI parser. Returns (project, name, version).
// An unparseable input yields the empty triple with no error so refresh paths
// can degrade gracefully on malformed server data.
func ParseEnvRef(ref string) (project, name, version string) {
	parts := strings.Split(ref, "/")
	switch len(parts) {
	case 1:
		project, name = "default", parts[0]
	case 2:
		project, name = parts[0], parts[1]
	default:
		// 3+ segments: <org>/<project>/<env>; PSP stores the env qualified
		// without the org prefix, so the org is the first segment we drop.
		project = parts[len(parts)-2]
		name = parts[len(parts)-1]
	}
	// Look for `@version` first, then fall back to `:version` for compat.
	if i := strings.LastIndex(name, "@"); i >= 0 {
		version = name[i+1:]
		name = name[:i]
	} else if i := strings.LastIndex(name, ":"); i >= 0 {
		version = name[i+1:]
		name = name[:i]
	}
	return project, name, version
}
