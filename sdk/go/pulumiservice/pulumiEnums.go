// Code generated by pulumigen DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package pulumiservice

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type TeamStackPermissionScope float64

const (
	// Grants read permissions to stack.
	TeamStackPermissionScopeRead = TeamStackPermissionScope(101)
	// Grants edit permissions to stack.
	TeamStackPermissionScopeEdit = TeamStackPermissionScope(102)
	// Grants admin permissions to stack.
	TeamStackPermissionScopeAdmin = TeamStackPermissionScope(103)
)

type WebhookFilters string

const (
	// Trigger a webhook when a stack is created. Only valid for org webhooks.
	WebhookFiltersStackCreated = WebhookFilters("stack_created")
	// Trigger a webhook when a stack is deleted. Only valid for org webhooks.
	WebhookFiltersStackDeleted = WebhookFilters("stack_deleted")
	// Trigger a webhook when a stack update succeeds.
	WebhookFiltersUpdateSucceeded = WebhookFilters("update_succeeded")
	// Trigger a webhook when a stack update fails.
	WebhookFiltersUpdateFailed = WebhookFilters("update_failed")
	// Trigger a webhook when a stack preview succeeds.
	WebhookFiltersPreviewSucceeded = WebhookFilters("preview_succeeded")
	// Trigger a webhook when a stack preview fails.
	WebhookFiltersPreviewFailed = WebhookFilters("preview_failed")
	// Trigger a webhook when a stack destroy succeeds.
	WebhookFiltersDestroySucceeded = WebhookFilters("destroy_succeeded")
	// Trigger a webhook when a stack destroy fails.
	WebhookFiltersDestroyFailed = WebhookFilters("destroy_failed")
	// Trigger a webhook when a stack refresh succeeds.
	WebhookFiltersRefreshSucceeded = WebhookFilters("refresh_succeeded")
	// Trigger a webhook when a stack refresh fails.
	WebhookFiltersRefreshFailed = WebhookFilters("refresh_failed")
	// Trigger a webhook when a deployment is queued.
	WebhookFiltersDeploymentQueued = WebhookFilters("deployment_queued")
	// Trigger a webhook when a deployment starts running.
	WebhookFiltersDeploymentStarted = WebhookFilters("deployment_started")
	// Trigger a webhook when a deployment succeeds.
	WebhookFiltersDeploymentSucceeded = WebhookFilters("deployment_succeeded")
	// Trigger a webhook when a deployment fails.
	WebhookFiltersDeploymentFailed = WebhookFilters("deployment_failed")
)

func (WebhookFilters) ElementType() reflect.Type {
	return reflect.TypeOf((*WebhookFilters)(nil)).Elem()
}

func (e WebhookFilters) ToWebhookFiltersOutput() WebhookFiltersOutput {
	return pulumi.ToOutput(e).(WebhookFiltersOutput)
}

func (e WebhookFilters) ToWebhookFiltersOutputWithContext(ctx context.Context) WebhookFiltersOutput {
	return pulumi.ToOutputWithContext(ctx, e).(WebhookFiltersOutput)
}

func (e WebhookFilters) ToWebhookFiltersPtrOutput() WebhookFiltersPtrOutput {
	return e.ToWebhookFiltersPtrOutputWithContext(context.Background())
}

func (e WebhookFilters) ToWebhookFiltersPtrOutputWithContext(ctx context.Context) WebhookFiltersPtrOutput {
	return WebhookFilters(e).ToWebhookFiltersOutputWithContext(ctx).ToWebhookFiltersPtrOutputWithContext(ctx)
}

func (e WebhookFilters) ToStringOutput() pulumi.StringOutput {
	return pulumi.ToOutput(pulumi.String(e)).(pulumi.StringOutput)
}

func (e WebhookFilters) ToStringOutputWithContext(ctx context.Context) pulumi.StringOutput {
	return pulumi.ToOutputWithContext(ctx, pulumi.String(e)).(pulumi.StringOutput)
}

func (e WebhookFilters) ToStringPtrOutput() pulumi.StringPtrOutput {
	return pulumi.String(e).ToStringPtrOutputWithContext(context.Background())
}

func (e WebhookFilters) ToStringPtrOutputWithContext(ctx context.Context) pulumi.StringPtrOutput {
	return pulumi.String(e).ToStringOutputWithContext(ctx).ToStringPtrOutputWithContext(ctx)
}

type WebhookFiltersOutput struct{ *pulumi.OutputState }

func (WebhookFiltersOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*WebhookFilters)(nil)).Elem()
}

func (o WebhookFiltersOutput) ToWebhookFiltersOutput() WebhookFiltersOutput {
	return o
}

func (o WebhookFiltersOutput) ToWebhookFiltersOutputWithContext(ctx context.Context) WebhookFiltersOutput {
	return o
}

func (o WebhookFiltersOutput) ToWebhookFiltersPtrOutput() WebhookFiltersPtrOutput {
	return o.ToWebhookFiltersPtrOutputWithContext(context.Background())
}

func (o WebhookFiltersOutput) ToWebhookFiltersPtrOutputWithContext(ctx context.Context) WebhookFiltersPtrOutput {
	return o.ApplyTWithContext(ctx, func(_ context.Context, v WebhookFilters) *WebhookFilters {
		return &v
	}).(WebhookFiltersPtrOutput)
}

func (o WebhookFiltersOutput) ToStringOutput() pulumi.StringOutput {
	return o.ToStringOutputWithContext(context.Background())
}

func (o WebhookFiltersOutput) ToStringOutputWithContext(ctx context.Context) pulumi.StringOutput {
	return o.ApplyTWithContext(ctx, func(_ context.Context, e WebhookFilters) string {
		return string(e)
	}).(pulumi.StringOutput)
}

func (o WebhookFiltersOutput) ToStringPtrOutput() pulumi.StringPtrOutput {
	return o.ToStringPtrOutputWithContext(context.Background())
}

func (o WebhookFiltersOutput) ToStringPtrOutputWithContext(ctx context.Context) pulumi.StringPtrOutput {
	return o.ApplyTWithContext(ctx, func(_ context.Context, e WebhookFilters) *string {
		v := string(e)
		return &v
	}).(pulumi.StringPtrOutput)
}

type WebhookFiltersPtrOutput struct{ *pulumi.OutputState }

func (WebhookFiltersPtrOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**WebhookFilters)(nil)).Elem()
}

func (o WebhookFiltersPtrOutput) ToWebhookFiltersPtrOutput() WebhookFiltersPtrOutput {
	return o
}

func (o WebhookFiltersPtrOutput) ToWebhookFiltersPtrOutputWithContext(ctx context.Context) WebhookFiltersPtrOutput {
	return o
}

func (o WebhookFiltersPtrOutput) Elem() WebhookFiltersOutput {
	return o.ApplyT(func(v *WebhookFilters) WebhookFilters {
		if v != nil {
			return *v
		}
		var ret WebhookFilters
		return ret
	}).(WebhookFiltersOutput)
}

func (o WebhookFiltersPtrOutput) ToStringPtrOutput() pulumi.StringPtrOutput {
	return o.ToStringPtrOutputWithContext(context.Background())
}

func (o WebhookFiltersPtrOutput) ToStringPtrOutputWithContext(ctx context.Context) pulumi.StringPtrOutput {
	return o.ApplyTWithContext(ctx, func(_ context.Context, e *WebhookFilters) *string {
		if e == nil {
			return nil
		}
		v := string(*e)
		return &v
	}).(pulumi.StringPtrOutput)
}

// WebhookFiltersInput is an input type that accepts WebhookFiltersArgs and WebhookFiltersOutput values.
// You can construct a concrete instance of `WebhookFiltersInput` via:
//
//	WebhookFiltersArgs{...}
type WebhookFiltersInput interface {
	pulumi.Input

	ToWebhookFiltersOutput() WebhookFiltersOutput
	ToWebhookFiltersOutputWithContext(context.Context) WebhookFiltersOutput
}

var webhookFiltersPtrType = reflect.TypeOf((**WebhookFilters)(nil)).Elem()

type WebhookFiltersPtrInput interface {
	pulumi.Input

	ToWebhookFiltersPtrOutput() WebhookFiltersPtrOutput
	ToWebhookFiltersPtrOutputWithContext(context.Context) WebhookFiltersPtrOutput
}

type webhookFiltersPtr string

func WebhookFiltersPtr(v string) WebhookFiltersPtrInput {
	return (*webhookFiltersPtr)(&v)
}

func (*webhookFiltersPtr) ElementType() reflect.Type {
	return webhookFiltersPtrType
}

func (in *webhookFiltersPtr) ToWebhookFiltersPtrOutput() WebhookFiltersPtrOutput {
	return pulumi.ToOutput(in).(WebhookFiltersPtrOutput)
}

func (in *webhookFiltersPtr) ToWebhookFiltersPtrOutputWithContext(ctx context.Context) WebhookFiltersPtrOutput {
	return pulumi.ToOutputWithContext(ctx, in).(WebhookFiltersPtrOutput)
}

// WebhookFiltersArrayInput is an input type that accepts WebhookFiltersArray and WebhookFiltersArrayOutput values.
// You can construct a concrete instance of `WebhookFiltersArrayInput` via:
//
//	WebhookFiltersArray{ WebhookFiltersArgs{...} }
type WebhookFiltersArrayInput interface {
	pulumi.Input

	ToWebhookFiltersArrayOutput() WebhookFiltersArrayOutput
	ToWebhookFiltersArrayOutputWithContext(context.Context) WebhookFiltersArrayOutput
}

type WebhookFiltersArray []WebhookFilters

func (WebhookFiltersArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]WebhookFilters)(nil)).Elem()
}

func (i WebhookFiltersArray) ToWebhookFiltersArrayOutput() WebhookFiltersArrayOutput {
	return i.ToWebhookFiltersArrayOutputWithContext(context.Background())
}

func (i WebhookFiltersArray) ToWebhookFiltersArrayOutputWithContext(ctx context.Context) WebhookFiltersArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(WebhookFiltersArrayOutput)
}

type WebhookFiltersArrayOutput struct{ *pulumi.OutputState }

func (WebhookFiltersArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]WebhookFilters)(nil)).Elem()
}

func (o WebhookFiltersArrayOutput) ToWebhookFiltersArrayOutput() WebhookFiltersArrayOutput {
	return o
}

func (o WebhookFiltersArrayOutput) ToWebhookFiltersArrayOutputWithContext(ctx context.Context) WebhookFiltersArrayOutput {
	return o
}

func (o WebhookFiltersArrayOutput) Index(i pulumi.IntInput) WebhookFiltersOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) WebhookFilters {
		return vs[0].([]WebhookFilters)[vs[1].(int)]
	}).(WebhookFiltersOutput)
}

type WebhookFormat string

const (
	// The default webhook format.
	WebhookFormatRaw = WebhookFormat("raw")
	// Messages formatted for consumption by Slack incoming webhooks.
	WebhookFormatSlack = WebhookFormat("slack")
	// Initiate deployments on a stack from a Pulumi Cloud webhook.
	WebhookFormatPulumiDeployments = WebhookFormat("pulumi_deployments")
	// Messages formatted for consumption by Microsoft Teams incoming webhooks.
	WebhookFormatMicrosoftTeams = WebhookFormat("ms_teams")
)

func (WebhookFormat) ElementType() reflect.Type {
	return reflect.TypeOf((*WebhookFormat)(nil)).Elem()
}

func (e WebhookFormat) ToWebhookFormatOutput() WebhookFormatOutput {
	return pulumi.ToOutput(e).(WebhookFormatOutput)
}

func (e WebhookFormat) ToWebhookFormatOutputWithContext(ctx context.Context) WebhookFormatOutput {
	return pulumi.ToOutputWithContext(ctx, e).(WebhookFormatOutput)
}

func (e WebhookFormat) ToWebhookFormatPtrOutput() WebhookFormatPtrOutput {
	return e.ToWebhookFormatPtrOutputWithContext(context.Background())
}

func (e WebhookFormat) ToWebhookFormatPtrOutputWithContext(ctx context.Context) WebhookFormatPtrOutput {
	return WebhookFormat(e).ToWebhookFormatOutputWithContext(ctx).ToWebhookFormatPtrOutputWithContext(ctx)
}

func (e WebhookFormat) ToStringOutput() pulumi.StringOutput {
	return pulumi.ToOutput(pulumi.String(e)).(pulumi.StringOutput)
}

func (e WebhookFormat) ToStringOutputWithContext(ctx context.Context) pulumi.StringOutput {
	return pulumi.ToOutputWithContext(ctx, pulumi.String(e)).(pulumi.StringOutput)
}

func (e WebhookFormat) ToStringPtrOutput() pulumi.StringPtrOutput {
	return pulumi.String(e).ToStringPtrOutputWithContext(context.Background())
}

func (e WebhookFormat) ToStringPtrOutputWithContext(ctx context.Context) pulumi.StringPtrOutput {
	return pulumi.String(e).ToStringOutputWithContext(ctx).ToStringPtrOutputWithContext(ctx)
}

type WebhookFormatOutput struct{ *pulumi.OutputState }

func (WebhookFormatOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*WebhookFormat)(nil)).Elem()
}

func (o WebhookFormatOutput) ToWebhookFormatOutput() WebhookFormatOutput {
	return o
}

func (o WebhookFormatOutput) ToWebhookFormatOutputWithContext(ctx context.Context) WebhookFormatOutput {
	return o
}

func (o WebhookFormatOutput) ToWebhookFormatPtrOutput() WebhookFormatPtrOutput {
	return o.ToWebhookFormatPtrOutputWithContext(context.Background())
}

func (o WebhookFormatOutput) ToWebhookFormatPtrOutputWithContext(ctx context.Context) WebhookFormatPtrOutput {
	return o.ApplyTWithContext(ctx, func(_ context.Context, v WebhookFormat) *WebhookFormat {
		return &v
	}).(WebhookFormatPtrOutput)
}

func (o WebhookFormatOutput) ToStringOutput() pulumi.StringOutput {
	return o.ToStringOutputWithContext(context.Background())
}

func (o WebhookFormatOutput) ToStringOutputWithContext(ctx context.Context) pulumi.StringOutput {
	return o.ApplyTWithContext(ctx, func(_ context.Context, e WebhookFormat) string {
		return string(e)
	}).(pulumi.StringOutput)
}

func (o WebhookFormatOutput) ToStringPtrOutput() pulumi.StringPtrOutput {
	return o.ToStringPtrOutputWithContext(context.Background())
}

func (o WebhookFormatOutput) ToStringPtrOutputWithContext(ctx context.Context) pulumi.StringPtrOutput {
	return o.ApplyTWithContext(ctx, func(_ context.Context, e WebhookFormat) *string {
		v := string(e)
		return &v
	}).(pulumi.StringPtrOutput)
}

type WebhookFormatPtrOutput struct{ *pulumi.OutputState }

func (WebhookFormatPtrOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**WebhookFormat)(nil)).Elem()
}

func (o WebhookFormatPtrOutput) ToWebhookFormatPtrOutput() WebhookFormatPtrOutput {
	return o
}

func (o WebhookFormatPtrOutput) ToWebhookFormatPtrOutputWithContext(ctx context.Context) WebhookFormatPtrOutput {
	return o
}

func (o WebhookFormatPtrOutput) Elem() WebhookFormatOutput {
	return o.ApplyT(func(v *WebhookFormat) WebhookFormat {
		if v != nil {
			return *v
		}
		var ret WebhookFormat
		return ret
	}).(WebhookFormatOutput)
}

func (o WebhookFormatPtrOutput) ToStringPtrOutput() pulumi.StringPtrOutput {
	return o.ToStringPtrOutputWithContext(context.Background())
}

func (o WebhookFormatPtrOutput) ToStringPtrOutputWithContext(ctx context.Context) pulumi.StringPtrOutput {
	return o.ApplyTWithContext(ctx, func(_ context.Context, e *WebhookFormat) *string {
		if e == nil {
			return nil
		}
		v := string(*e)
		return &v
	}).(pulumi.StringPtrOutput)
}

// WebhookFormatInput is an input type that accepts WebhookFormatArgs and WebhookFormatOutput values.
// You can construct a concrete instance of `WebhookFormatInput` via:
//
//	WebhookFormatArgs{...}
type WebhookFormatInput interface {
	pulumi.Input

	ToWebhookFormatOutput() WebhookFormatOutput
	ToWebhookFormatOutputWithContext(context.Context) WebhookFormatOutput
}

var webhookFormatPtrType = reflect.TypeOf((**WebhookFormat)(nil)).Elem()

type WebhookFormatPtrInput interface {
	pulumi.Input

	ToWebhookFormatPtrOutput() WebhookFormatPtrOutput
	ToWebhookFormatPtrOutputWithContext(context.Context) WebhookFormatPtrOutput
}

type webhookFormatPtr string

func WebhookFormatPtr(v string) WebhookFormatPtrInput {
	return (*webhookFormatPtr)(&v)
}

func (*webhookFormatPtr) ElementType() reflect.Type {
	return webhookFormatPtrType
}

func (in *webhookFormatPtr) ToWebhookFormatPtrOutput() WebhookFormatPtrOutput {
	return pulumi.ToOutput(in).(WebhookFormatPtrOutput)
}

func (in *webhookFormatPtr) ToWebhookFormatPtrOutputWithContext(ctx context.Context) WebhookFormatPtrOutput {
	return pulumi.ToOutputWithContext(ctx, in).(WebhookFormatPtrOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*WebhookFiltersInput)(nil)).Elem(), WebhookFilters("stack_created"))
	pulumi.RegisterInputType(reflect.TypeOf((*WebhookFiltersPtrInput)(nil)).Elem(), WebhookFilters("stack_created"))
	pulumi.RegisterInputType(reflect.TypeOf((*WebhookFiltersArrayInput)(nil)).Elem(), WebhookFiltersArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*WebhookFormatInput)(nil)).Elem(), WebhookFormat("raw"))
	pulumi.RegisterInputType(reflect.TypeOf((*WebhookFormatPtrInput)(nil)).Elem(), WebhookFormat("raw"))
	pulumi.RegisterOutputType(WebhookFiltersOutput{})
	pulumi.RegisterOutputType(WebhookFiltersPtrOutput{})
	pulumi.RegisterOutputType(WebhookFiltersArrayOutput{})
	pulumi.RegisterOutputType(WebhookFormatOutput{})
	pulumi.RegisterOutputType(WebhookFormatPtrOutput{})
}
