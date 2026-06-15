// Copyright 2026, Pulumi Corporation.
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

package resources

import (
	"context"
	"fmt"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

// defaultWebhookGroups maps a webhook scope (organization/stack/environment)
// to the default set of event-group subscriptions applied when the user has
// not provided either `filters` or `groups`.
var defaultWebhookGroups = map[string][]WebhookGroup{
	"organization": {WebhookGroupDeployments, WebhookGroupEnvironments, WebhookGroupStacks},
	"stack":        {WebhookGroupDeployments, WebhookGroupStacks},
	"environment":  {WebhookGroupEnvironments},
}

// WebhookFormat is the wire format of webhook payloads.
type WebhookFormat string

const (
	WebhookFormatRaw               WebhookFormat = "raw"
	WebhookFormatSlack             WebhookFormat = "slack"
	WebhookFormatPulumiDeployments WebhookFormat = "pulumi_deployments"
	WebhookFormatMicrosoftTeams    WebhookFormat = "ms_teams"
)

func (WebhookFormat) Values() []infer.EnumValue[WebhookFormat] {
	return []infer.EnumValue[WebhookFormat]{
		{Value: WebhookFormatRaw, Description: "The default webhook format."},
		{
			Value:       WebhookFormatSlack,
			Description: "Messages formatted for consumption by Slack incoming webhooks.",
		},
		{
			Name:        "PulumiDeployments",
			Value:       WebhookFormatPulumiDeployments,
			Description: "Initiate deployments on a stack from a Pulumi Cloud webhook.",
		},
		{
			Name:        "MicrosoftTeams",
			Value:       WebhookFormatMicrosoftTeams,
			Description: "Messages formatted for consumption by Microsoft Teams incoming webhooks.",
		},
	}
}

// WebhookGroup is a high-level bundle of webhook event filters.
type WebhookGroup string

const (
	WebhookGroupStacks       WebhookGroup = "stacks"
	WebhookGroupDeployments  WebhookGroup = "deployments"
	WebhookGroupEnvironments WebhookGroup = "environments"
)

func (WebhookGroup) Values() []infer.EnumValue[WebhookGroup] {
	return []infer.EnumValue[WebhookGroup]{
		{Name: "Stacks", Value: WebhookGroupStacks, Description: "A group of webhooks containing all stack events."},
		{
			Name:        "Deployments",
			Value:       WebhookGroupDeployments,
			Description: "A group of webhooks containing all deployment events.",
		},
		{
			Name:        "Environments",
			Value:       WebhookGroupEnvironments,
			Description: "A group of webhooks containing all environment events.",
		},
	}
}

// WebhookFilters is the set of fine-grained event filters that may be
// subscribed to.
type WebhookFilters string

const (
	WebhookFilterStackCreated                 WebhookFilters = "stack_created"
	WebhookFilterStackDeleted                 WebhookFilters = "stack_deleted"
	WebhookFilterUpdateSucceeded              WebhookFilters = "update_succeeded"
	WebhookFilterUpdateFailed                 WebhookFilters = "update_failed"
	WebhookFilterPreviewSucceeded             WebhookFilters = "preview_succeeded"
	WebhookFilterPreviewFailed                WebhookFilters = "preview_failed"
	WebhookFilterDestroySucceeded             WebhookFilters = "destroy_succeeded"
	WebhookFilterDestroyFailed                WebhookFilters = "destroy_failed"
	WebhookFilterRefreshSucceeded             WebhookFilters = "refresh_succeeded"
	WebhookFilterRefreshFailed                WebhookFilters = "refresh_failed"
	WebhookFilterDeploymentQueued             WebhookFilters = "deployment_queued"
	WebhookFilterDeploymentStarted            WebhookFilters = "deployment_started"
	WebhookFilterDeploymentSucceeded          WebhookFilters = "deployment_succeeded"
	WebhookFilterDeploymentFailed             WebhookFilters = "deployment_failed"
	WebhookFilterDriftDetected                WebhookFilters = "drift_detected"
	WebhookFilterDriftDetectionSucceeded      WebhookFilters = "drift_detection_succeeded"
	WebhookFilterDriftDetectionFailed         WebhookFilters = "drift_detection_failed"
	WebhookFilterDriftRemediationSucceeded    WebhookFilters = "drift_remediation_succeeded"
	WebhookFilterDriftRemediationFailed       WebhookFilters = "drift_remediation_failed"
	WebhookFilterEnvironmentCreated           WebhookFilters = "environment_created"
	WebhookFilterEnvironmentDeleted           WebhookFilters = "environment_deleted"
	WebhookFilterEnvironmentRevisionCreated   WebhookFilters = "environment_revision_created"
	WebhookFilterEnvRevisionRetracted         WebhookFilters = "environment_revision_retracted"
	WebhookFilterEnvRevisionTagCreated        WebhookFilters = "environment_revision_tag_created"
	WebhookFilterEnvRevisionTagDeleted        WebhookFilters = "environment_revision_tag_deleted"
	WebhookFilterEnvRevisionTagUpdated        WebhookFilters = "environment_revision_tag_updated"
	WebhookFilterEnvironmentTagCreated        WebhookFilters = "environment_tag_created"
	WebhookFilterEnvironmentTagDeleted        WebhookFilters = "environment_tag_deleted"
	WebhookFilterEnvironmentTagUpdated        WebhookFilters = "environment_tag_updated"
	WebhookFilterImportedEnvironmentChanged   WebhookFilters = "imported_environment_changed"
	WebhookFilterEnvironmentRotationSucceeded WebhookFilters = "environment_rotation_succeeded"
	WebhookFilterEnvironmentRotationFailed    WebhookFilters = "environment_rotation_failed"
)

func (WebhookFilters) Values() []infer.EnumValue[WebhookFilters] {
	return []infer.EnumValue[WebhookFilters]{
		{
			Name:        "StackCreated",
			Value:       WebhookFilterStackCreated,
			Description: "Trigger a webhook when a stack is created. Only valid for org webhooks.",
		},
		{
			Name:        "StackDeleted",
			Value:       WebhookFilterStackDeleted,
			Description: "Trigger a webhook when a stack is deleted. Only valid for org webhooks.",
		},
		{
			Name:        "UpdateSucceeded",
			Value:       WebhookFilterUpdateSucceeded,
			Description: "Trigger a webhook when a stack update succeeds.",
		},
		{
			Name:        "UpdateFailed",
			Value:       WebhookFilterUpdateFailed,
			Description: "Trigger a webhook when a stack update fails.",
		},
		{
			Name:        "PreviewSucceeded",
			Value:       WebhookFilterPreviewSucceeded,
			Description: "Trigger a webhook when a stack preview succeeds.",
		},
		{
			Name:        "PreviewFailed",
			Value:       WebhookFilterPreviewFailed,
			Description: "Trigger a webhook when a stack preview fails.",
		},
		{
			Name:        "DestroySucceeded",
			Value:       WebhookFilterDestroySucceeded,
			Description: "Trigger a webhook when a stack destroy succeeds.",
		},
		{
			Name:        "DestroyFailed",
			Value:       WebhookFilterDestroyFailed,
			Description: "Trigger a webhook when a stack destroy fails.",
		},
		{
			Name:        "RefreshSucceeded",
			Value:       WebhookFilterRefreshSucceeded,
			Description: "Trigger a webhook when a stack refresh succeeds.",
		},
		{
			Name:        "RefreshFailed",
			Value:       WebhookFilterRefreshFailed,
			Description: "Trigger a webhook when a stack refresh fails.",
		},
		{
			Name:        "DeploymentQueued",
			Value:       WebhookFilterDeploymentQueued,
			Description: "Trigger a webhook when a deployment is queued.",
		},
		{
			Name:        "DeploymentStarted",
			Value:       WebhookFilterDeploymentStarted,
			Description: "Trigger a webhook when a deployment starts running.",
		},
		{
			Name:        "DeploymentSucceeded",
			Value:       WebhookFilterDeploymentSucceeded,
			Description: "Trigger a webhook when a deployment succeeds.",
		},
		{
			Name:        "DeploymentFailed",
			Value:       WebhookFilterDeploymentFailed,
			Description: "Trigger a webhook when a deployment fails.",
		},
		{
			Name:        "DriftDetected",
			Value:       WebhookFilterDriftDetected,
			Description: "Trigger a webhook when drift is detected.",
		},
		{
			Name:  "DriftDetectionSucceeded",
			Value: WebhookFilterDriftDetectionSucceeded,
			Description: "Trigger a webhook when a drift detection run succeeds, " +
				"regardless of whether drift is detected.",
		},
		{
			Name:        "DriftDetectionFailed",
			Value:       WebhookFilterDriftDetectionFailed,
			Description: "Trigger a webhook when a drift detection run fails.",
		},
		{
			Name:        "DriftRemediationSucceeded",
			Value:       WebhookFilterDriftRemediationSucceeded,
			Description: "Trigger a webhook when a drift remediation run succeeds.",
		},
		{
			Name:        "DriftRemediationFailed",
			Value:       WebhookFilterDriftRemediationFailed,
			Description: "Trigger a webhook when a drift remediation run fails.",
		},
		{
			Name:        "EnvironmentCreated",
			Value:       WebhookFilterEnvironmentCreated,
			Description: "Trigger a webhook when a new environment is created.",
		},
		{
			Name:        "EnvironmentDeleted",
			Value:       WebhookFilterEnvironmentDeleted,
			Description: "Trigger a webhook when an environment is deleted.",
		},
		{
			Name:        "EnvironmentRevisionCreated",
			Value:       WebhookFilterEnvironmentRevisionCreated,
			Description: "Trigger a webhook when a new revision is created on an environment.",
		},
		{
			Name:        "EnvironmentRevisionRetracted",
			Value:       WebhookFilterEnvRevisionRetracted,
			Description: "Trigger a webhook when a revision is retracted on an environment.",
		},
		{
			Name:        "EnvironmentRevisionTagCreated",
			Value:       WebhookFilterEnvRevisionTagCreated,
			Description: "Trigger a webhook when a revision tag is created on an environment.",
		},
		{
			Name:        "EnvironmentRevisionTagDeleted",
			Value:       WebhookFilterEnvRevisionTagDeleted,
			Description: "Trigger a webhook when a revision tag is deleted on an environment.",
		},
		{
			Name:        "EnvironmentRevisionTagUpdated",
			Value:       WebhookFilterEnvRevisionTagUpdated,
			Description: "Trigger a webhook when a revision tag is updated on an environment.",
		},
		{
			Name:        "EnvironmentTagCreated",
			Value:       WebhookFilterEnvironmentTagCreated,
			Description: "Trigger a webhook when an environment tag is created.",
		},
		{
			Name:        "EnvironmentTagDeleted",
			Value:       WebhookFilterEnvironmentTagDeleted,
			Description: "Trigger a webhook when an environment tag is deleted.",
		},
		{
			Name:        "EnvironmentTagUpdated",
			Value:       WebhookFilterEnvironmentTagUpdated,
			Description: "Trigger a webhook when an environment tag is updated.",
		},
		{
			Name:        "ImportedEnvironmentChanged",
			Value:       WebhookFilterImportedEnvironmentChanged,
			Description: "Trigger a webhook when an imported environment has changed.",
		},
		{
			Name:        "EnvironmentRotationSucceeded",
			Value:       WebhookFilterEnvironmentRotationSucceeded,
			Description: "Trigger a webhook when an environment rotation succeeds.",
		},
		{
			Name:        "EnvironmentRotationFailed",
			Value:       WebhookFilterEnvironmentRotationFailed,
			Description: "Trigger a webhook when an environment rotation fails.",
		},
	}
}

type Webhook struct{}

var (
	_ infer.CustomCheck[WebhookInput]                = &Webhook{}
	_ infer.CustomCreate[WebhookInput, WebhookState] = &Webhook{}
	_ infer.CustomUpdate[WebhookInput, WebhookState] = &Webhook{}
	_ infer.CustomDelete[WebhookState]               = &Webhook{}
	_ infer.CustomRead[WebhookInput, WebhookState]   = &Webhook{}
)

func (*Webhook) Annotate(a infer.Annotator) {
	a.Describe(&Webhook{},
		"Pulumi Webhooks allow you to notify external services of events happening within your Pulumi "+
			"organization or stack. For example, you can trigger a notification whenever a stack is updated. "+
			"Whenever an event occurs, Pulumi will send an HTTP POST request to all registered webhooks. The "+
			"webhook can then be used to emit some notification, start running integration tests, or even "+
			"update additional stacks.\n\n### Import\n\nPulumi webhooks can be imported using the `id`, which "+
			"for webhooks is `{org}/{project}/{stack}/{webhook-name}` e.g.,\n\n```sh\n $ pulumi import "+
			"pulumiservice:index:Webhook my_webhook my-org/my-project/my-stack/4b0d0671\n```\n\n")
	a.SetToken("index", "Webhook")
}

type WebhookInput struct {
	Active           bool             `pulumi:"active"`
	DisplayName      string           `pulumi:"displayName"`
	PayloadURL       string           `pulumi:"payloadUrl"`
	Secret           *string          `pulumi:"secret,optional"          provider:"secret"`
	OrganizationName string           `pulumi:"organizationName"         provider:"replaceOnChanges"`
	ProjectName      *string          `pulumi:"projectName,optional"     provider:"replaceOnChanges"`
	StackName        *string          `pulumi:"stackName,optional"       provider:"replaceOnChanges"`
	EnvironmentName  *string          `pulumi:"environmentName,optional" provider:"replaceOnChanges"`
	Format           *WebhookFormat   `pulumi:"format,optional"`
	Filters          []WebhookFilters `pulumi:"filters,optional"`
	Groups           []WebhookGroup   `pulumi:"groups,optional"`
}

func (i *WebhookInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Active, "Indicates whether this webhook is enabled or not.")
	a.Describe(&i.DisplayName, "The friendly name displayed in the Pulumi Cloud.")
	a.Describe(&i.PayloadURL, "URL to send request to.")
	a.Describe(
		&i.Secret,
		"Optional. secret used as the HMAC key. See [webhook docs]"+
			"(https://www.pulumi.com/docs/intro/pulumi-service/webhooks/#headers) for more information.",
	)
	a.Describe(&i.OrganizationName, "Name of the organization.")
	a.Describe(&i.ProjectName, "Name of the project. Only specified if this is a stack or environment webhook.")
	a.Describe(&i.StackName, "Name of the stack. Only needed if this is a stack webhook.")
	a.Describe(&i.EnvironmentName, "Name of the environment. Only specified if this is an environment webhook.")
	a.Describe(
		&i.Format,
		"Format of the webhook payload. Can be either `raw`, `slack`, `ms_teams` or `pulumi_deployments`. "+
			"Defaults to `raw`.",
	)
	a.SetDefault(&i.Format, WebhookFormatRaw)
	a.Describe(
		&i.Filters,
		"Optional set of filters to apply to the webhook. See [webhook docs]"+
			"(https://www.pulumi.com/docs/intro/pulumi-service/webhooks/#filters) for more information.",
	)
	a.Describe(
		&i.Groups,
		"Optional set of filter groups to apply to the webhook. See [webhook docs]"+
			"(https://www.pulumi.com/docs/intro/pulumi-service/webhooks/#groups) for more information.",
	)
}

type WebhookState struct {
	WebhookInput
	// Format shadows the optional WebhookInput.Format so the output is marked
	// required: it always has a value (the input defaults to `raw`), so on the
	// output side it is never absent.
	//
	// TODO[github.com/pulumi/pulumi-go-provider/issues/537]: once infer marks
	// defaulted inputs as required outputs, drop this field and populate the
	// embedded WebhookInput.Format directly.
	Format WebhookFormat `pulumi:"format"`
	Name   string        `pulumi:"name"`
}

func (s *WebhookState) Annotate(a infer.Annotator) {
	a.Describe(
		&s.Format,
		"Format of the webhook payload. Can be either `raw`, `slack`, `ms_teams` or `pulumi_deployments`. "+
			"Defaults to `raw`.",
	)
	a.Describe(&s.Name, "Webhook identifier generated by Pulumi Cloud.")
}

// Check validates webhook scope (organization/stack/environment) inputs and
// applies default groups when neither filters nor groups are configured.
func (*Webhook) Check(
	ctx context.Context, req infer.CheckRequest,
) (infer.CheckResponse[WebhookInput], error) {
	i, failures, err := infer.DefaultCheck[WebhookInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[WebhookInput]{}, err
	}

	hasProject := i.ProjectName != nil
	hasStack := i.StackName != nil
	hasEnv := i.EnvironmentName != nil

	if hasStack && !hasProject {
		failures = append(failures, p.CheckFailure{
			Property: "projectName",
			Reason:   "projectName and stackName must both be specified for stack webhooks",
		})
	}
	if hasEnv && !hasProject {
		failures = append(failures, p.CheckFailure{
			Property: "projectName",
			Reason:   "projectName and environmentName must both be specified for environment webhooks",
		})
	}
	if hasEnv && hasStack {
		failures = append(failures, p.CheckFailure{
			Property: "stackName",
			Reason: "stackName needs to be empty if this is meant to be an environment webhook; " +
				"environmentName needs to be empty if this is meant to be a stack webhook",
		})
	}
	if hasProject && !hasStack && !hasEnv {
		failures = append(failures, p.CheckFailure{
			Property: "projectName",
			Reason: "projectName needs to be empty if this is meant to be an organization webhook; " +
				"otherwise provide stackName for stack webhook or environmentName for environment webhook",
		})
	}

	// If neither filters nor groups are specified, apply the scope-appropriate
	// default group set. This matches legacy provider behavior.
	if len(i.Filters) == 0 && len(i.Groups) == 0 {
		scope := "organization"
		if hasStack {
			scope = "stack"
		} else if hasEnv {
			scope = "environment"
		}
		i.Groups = append([]WebhookGroup(nil), defaultWebhookGroups[scope]...)
	}

	return infer.CheckResponse[WebhookInput]{Inputs: i, Failures: failures}, nil
}

func (*Webhook) Create(
	ctx context.Context, req infer.CreateRequest[WebhookInput],
) (infer.CreateResponse[WebhookState], error) {
	if req.DryRun {
		return infer.CreateResponse[WebhookState]{
			Output: WebhookState{WebhookInput: req.Inputs, Format: *req.Inputs.Format},
		}, nil
	}

	apiReq := toWebhookRequest(req.Inputs)
	webhook, err := config.GetClient(ctx).CreateWebhook(ctx, apiReq)
	if err != nil {
		return infer.CreateResponse[WebhookState]{}, fmt.Errorf(
			"error creating webhook %q: %w", req.Inputs.DisplayName, err,
		)
	}

	return infer.CreateResponse[WebhookState]{
		ID: generateWebhookID(req.Inputs, webhook.Name),
		Output: WebhookState{
			WebhookInput: req.Inputs,
			Format:       *req.Inputs.Format,
			Name:         webhook.Name,
		},
	}, nil
}

func (*Webhook) Update(
	ctx context.Context, req infer.UpdateRequest[WebhookInput, WebhookState],
) (infer.UpdateResponse[WebhookState], error) {
	if req.DryRun {
		return infer.UpdateResponse[WebhookState]{
			Output: WebhookState{
				WebhookInput: req.Inputs,
				Format:       *req.Inputs.Format,
				Name:         req.State.Name,
			},
		}, nil
	}

	updateReq := pulumiapi.UpdateWebhookRequest{
		WebhookRequest: toWebhookRequest(req.Inputs),
		Name:           req.State.Name,
	}
	if _, err := config.GetClient(ctx).UpdateWebhook(ctx, updateReq); err != nil {
		return infer.UpdateResponse[WebhookState]{}, fmt.Errorf(
			"error updating webhook %q: %w", req.State.Name, err,
		)
	}
	return infer.UpdateResponse[WebhookState]{
		Output: WebhookState{
			WebhookInput: req.Inputs,
			Format:       *req.Inputs.Format,
			Name:         req.State.Name,
		},
	}, nil
}

func (*Webhook) Delete(
	ctx context.Context, req infer.DeleteRequest[WebhookState],
) (infer.DeleteResponse, error) {
	hookID, err := splitWebhookID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteWebhook(
		ctx,
		hookID.organizationName,
		hookID.projectName,
		hookID.stackName,
		hookID.environmentName,
		hookID.webhookName,
	)
}

func (*Webhook) Read(
	ctx context.Context, req infer.ReadRequest[WebhookInput, WebhookState],
) (infer.ReadResponse[WebhookInput, WebhookState], error) {
	hookID, err := splitWebhookID(req.ID)
	if err != nil {
		return infer.ReadResponse[WebhookInput, WebhookState]{}, err
	}

	webhook, err := config.GetClient(ctx).GetWebhook(
		ctx,
		hookID.organizationName,
		hookID.projectName,
		hookID.stackName,
		hookID.environmentName,
		hookID.webhookName,
	)
	if err != nil {
		return infer.ReadResponse[WebhookInput, WebhookState]{}, fmt.Errorf(
			"failed to read webhook %q: %w", req.ID, err,
		)
	}
	if webhook == nil {
		return infer.ReadResponse[WebhookInput, WebhookState]{}, nil
	}

	format := WebhookFormat(webhook.Format)
	inputs := WebhookInput{
		Active:           webhook.Active,
		DisplayName:      webhook.DisplayName,
		PayloadURL:       webhook.PayloadURL,
		OrganizationName: hookID.organizationName,
		ProjectName:      hookID.projectName,
		StackName:        hookID.stackName,
		EnvironmentName:  hookID.environmentName,
		Format:           &format,
		Filters:          toWebhookFilters(webhook.Filters),
		Groups:           toWebhookGroups(webhook.Groups),
		// The API never returns the plaintext secret. Preserve the value
		// previously persisted in state so refresh does not erase it.
		Secret: req.State.Secret,
	}

	return infer.ReadResponse[WebhookInput, WebhookState]{
		ID:     req.ID,
		Inputs: inputs,
		State: WebhookState{
			WebhookInput: inputs,
			Format:       format,
			Name:         webhook.Name,
		},
	}, nil
}

// toWebhookRequest builds a pulumiapi request from infer inputs.
func toWebhookRequest(i WebhookInput) pulumiapi.WebhookRequest {
	req := pulumiapi.WebhookRequest{
		OrganizationName: i.OrganizationName,
		ProjectName:      i.ProjectName,
		StackName:        i.StackName,
		EnvironmentName:  i.EnvironmentName,
		DisplayName:      i.DisplayName,
		PayloadURL:       i.PayloadURL,
		Secret:           i.Secret,
		Active:           i.Active,
		Filters:          fromWebhookFilters(i.Filters),
		Groups:           fromWebhookGroups(i.Groups),
	}
	if i.Format != nil {
		s := string(*i.Format)
		req.Format = &s
	}
	return req
}

func toWebhookFilters(in []string) []WebhookFilters {
	if len(in) == 0 {
		return nil
	}
	out := make([]WebhookFilters, len(in))
	for idx, v := range in {
		out[idx] = WebhookFilters(v)
	}
	return out
}

func fromWebhookFilters(in []WebhookFilters) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	for idx, v := range in {
		out[idx] = string(v)
	}
	return out
}

func toWebhookGroups(in []string) []WebhookGroup {
	if len(in) == 0 {
		return nil
	}
	out := make([]WebhookGroup, len(in))
	for idx, v := range in {
		out[idx] = WebhookGroup(v)
	}
	return out
}

func fromWebhookGroups(in []WebhookGroup) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	for idx, v := range in {
		out[idx] = string(v)
	}
	return out
}

func generateWebhookID(input WebhookInput, webhookName string) string {
	if input.ProjectName != nil && input.StackName != nil {
		return fmt.Sprintf("%s/%s/%s/%s", input.OrganizationName, *input.ProjectName, *input.StackName, webhookName)
	}
	if input.ProjectName != nil && input.EnvironmentName != nil {
		// The "environment" segment disambiguates this from stack webhooks,
		// which also have four slash-separated parts.
		return fmt.Sprintf(
			"%s/environment/%s/%s/%s",
			input.OrganizationName, *input.ProjectName, *input.EnvironmentName, webhookName,
		)
	}
	return fmt.Sprintf("%s/%s", input.OrganizationName, webhookName)
}

type webhookID struct {
	organizationName string
	projectName      *string
	stackName        *string
	environmentName  *string
	webhookName      string
}

func splitWebhookID(id string) (*webhookID, error) {
	// Accepted formats:
	//   organization/webhookName                                       (org)
	//   organization/project/stack/webhookName                         (stack)
	//   organization/environment/project/environmentName/webhookName   (env)
	s := strings.Split(id, "/")
	switch len(s) {
	case 2:
		return &webhookID{
			organizationName: s[0],
			webhookName:      s[1],
		}, nil
	case 4:
		return &webhookID{
			organizationName: s[0],
			projectName:      &s[1],
			stackName:        &s[2],
			webhookName:      s[3],
		}, nil
	case 5:
		return &webhookID{
			organizationName: s[0],
			projectName:      &s[2],
			environmentName:  &s[3],
			webhookName:      s[4],
		}, nil
	default:
		return nil, fmt.Errorf("%q is not a valid webhook ID", id)
	}
}
