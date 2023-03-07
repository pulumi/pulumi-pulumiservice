// Code generated by pulumigen DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package pulumiservice

import (
	"context"
	"reflect"

	"errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The Pulumi Service allows users to create access tokens scoped to team. Team access tokens is a resource to create them and assign them to a team
type TeamAccessToken struct {
	pulumi.CustomResourceState

	// Optional. Description for the token.
	Description pulumi.StringPtrOutput `pulumi:"description"`
	// The name for the token. This must be unique amongst all machine tokens within your organization.
	Name pulumi.StringPtrOutput `pulumi:"name"`
	// The organization's name.
	OrganizationName pulumi.StringPtrOutput `pulumi:"organizationName"`
	// The team name.
	TeamName pulumi.StringPtrOutput `pulumi:"teamName"`
}

// NewTeamAccessToken registers a new resource with the given unique name, arguments, and options.
func NewTeamAccessToken(ctx *pulumi.Context,
	name string, args *TeamAccessTokenArgs, opts ...pulumi.ResourceOption) (*TeamAccessToken, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.Name == nil {
		return nil, errors.New("invalid value for required argument 'Name'")
	}
	if args.OrganizationName == nil {
		return nil, errors.New("invalid value for required argument 'OrganizationName'")
	}
	if args.TeamName == nil {
		return nil, errors.New("invalid value for required argument 'TeamName'")
	}
	var resource TeamAccessToken
	err := ctx.RegisterResource("pulumiservice:index:TeamAccessToken", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetTeamAccessToken gets an existing TeamAccessToken resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetTeamAccessToken(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *TeamAccessTokenState, opts ...pulumi.ResourceOption) (*TeamAccessToken, error) {
	var resource TeamAccessToken
	err := ctx.ReadResource("pulumiservice:index:TeamAccessToken", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering TeamAccessToken resources.
type teamAccessTokenState struct {
}

type TeamAccessTokenState struct {
}

func (TeamAccessTokenState) ElementType() reflect.Type {
	return reflect.TypeOf((*teamAccessTokenState)(nil)).Elem()
}

type teamAccessTokenArgs struct {
	// Optional. Team description.
	Description *string `pulumi:"description"`
	// The name for the token. This must be unique amongst all machine tokens within your organization.
	Name string `pulumi:"name"`
	// The organization's name.
	OrganizationName string `pulumi:"organizationName"`
	// The team name.
	TeamName string `pulumi:"teamName"`
}

// The set of arguments for constructing a TeamAccessToken resource.
type TeamAccessTokenArgs struct {
	// Optional. Team description.
	Description pulumi.StringPtrInput
	// The name for the token. This must be unique amongst all machine tokens within your organization.
	Name pulumi.StringInput
	// The organization's name.
	OrganizationName pulumi.StringInput
	// The team name.
	TeamName pulumi.StringInput
}

func (TeamAccessTokenArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*teamAccessTokenArgs)(nil)).Elem()
}

type TeamAccessTokenInput interface {
	pulumi.Input

	ToTeamAccessTokenOutput() TeamAccessTokenOutput
	ToTeamAccessTokenOutputWithContext(ctx context.Context) TeamAccessTokenOutput
}

func (*TeamAccessToken) ElementType() reflect.Type {
	return reflect.TypeOf((**TeamAccessToken)(nil)).Elem()
}

func (i *TeamAccessToken) ToTeamAccessTokenOutput() TeamAccessTokenOutput {
	return i.ToTeamAccessTokenOutputWithContext(context.Background())
}

func (i *TeamAccessToken) ToTeamAccessTokenOutputWithContext(ctx context.Context) TeamAccessTokenOutput {
	return pulumi.ToOutputWithContext(ctx, i).(TeamAccessTokenOutput)
}

// TeamAccessTokenArrayInput is an input type that accepts TeamAccessTokenArray and TeamAccessTokenArrayOutput values.
// You can construct a concrete instance of `TeamAccessTokenArrayInput` via:
//
//	TeamAccessTokenArray{ TeamAccessTokenArgs{...} }
type TeamAccessTokenArrayInput interface {
	pulumi.Input

	ToTeamAccessTokenArrayOutput() TeamAccessTokenArrayOutput
	ToTeamAccessTokenArrayOutputWithContext(context.Context) TeamAccessTokenArrayOutput
}

type TeamAccessTokenArray []TeamAccessTokenInput

func (TeamAccessTokenArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*TeamAccessToken)(nil)).Elem()
}

func (i TeamAccessTokenArray) ToTeamAccessTokenArrayOutput() TeamAccessTokenArrayOutput {
	return i.ToTeamAccessTokenArrayOutputWithContext(context.Background())
}

func (i TeamAccessTokenArray) ToTeamAccessTokenArrayOutputWithContext(ctx context.Context) TeamAccessTokenArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(TeamAccessTokenArrayOutput)
}

// TeamAccessTokenMapInput is an input type that accepts TeamAccessTokenMap and TeamAccessTokenMapOutput values.
// You can construct a concrete instance of `TeamAccessTokenMapInput` via:
//
//	TeamAccessTokenMap{ "key": TeamAccessTokenArgs{...} }
type TeamAccessTokenMapInput interface {
	pulumi.Input

	ToTeamAccessTokenMapOutput() TeamAccessTokenMapOutput
	ToTeamAccessTokenMapOutputWithContext(context.Context) TeamAccessTokenMapOutput
}

type TeamAccessTokenMap map[string]TeamAccessTokenInput

func (TeamAccessTokenMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*TeamAccessToken)(nil)).Elem()
}

func (i TeamAccessTokenMap) ToTeamAccessTokenMapOutput() TeamAccessTokenMapOutput {
	return i.ToTeamAccessTokenMapOutputWithContext(context.Background())
}

func (i TeamAccessTokenMap) ToTeamAccessTokenMapOutputWithContext(ctx context.Context) TeamAccessTokenMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(TeamAccessTokenMapOutput)
}

type TeamAccessTokenOutput struct{ *pulumi.OutputState }

func (TeamAccessTokenOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**TeamAccessToken)(nil)).Elem()
}

func (o TeamAccessTokenOutput) ToTeamAccessTokenOutput() TeamAccessTokenOutput {
	return o
}

func (o TeamAccessTokenOutput) ToTeamAccessTokenOutputWithContext(ctx context.Context) TeamAccessTokenOutput {
	return o
}

// Optional. Description for the token.
func (o TeamAccessTokenOutput) Description() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *TeamAccessToken) pulumi.StringPtrOutput { return v.Description }).(pulumi.StringPtrOutput)
}

// The name for the token. This must be unique amongst all machine tokens within your organization.
func (o TeamAccessTokenOutput) Name() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *TeamAccessToken) pulumi.StringPtrOutput { return v.Name }).(pulumi.StringPtrOutput)
}

// The organization's name.
func (o TeamAccessTokenOutput) OrganizationName() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *TeamAccessToken) pulumi.StringPtrOutput { return v.OrganizationName }).(pulumi.StringPtrOutput)
}

// The team name.
func (o TeamAccessTokenOutput) TeamName() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *TeamAccessToken) pulumi.StringPtrOutput { return v.TeamName }).(pulumi.StringPtrOutput)
}

type TeamAccessTokenArrayOutput struct{ *pulumi.OutputState }

func (TeamAccessTokenArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*TeamAccessToken)(nil)).Elem()
}

func (o TeamAccessTokenArrayOutput) ToTeamAccessTokenArrayOutput() TeamAccessTokenArrayOutput {
	return o
}

func (o TeamAccessTokenArrayOutput) ToTeamAccessTokenArrayOutputWithContext(ctx context.Context) TeamAccessTokenArrayOutput {
	return o
}

func (o TeamAccessTokenArrayOutput) Index(i pulumi.IntInput) TeamAccessTokenOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *TeamAccessToken {
		return vs[0].([]*TeamAccessToken)[vs[1].(int)]
	}).(TeamAccessTokenOutput)
}

type TeamAccessTokenMapOutput struct{ *pulumi.OutputState }

func (TeamAccessTokenMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*TeamAccessToken)(nil)).Elem()
}

func (o TeamAccessTokenMapOutput) ToTeamAccessTokenMapOutput() TeamAccessTokenMapOutput {
	return o
}

func (o TeamAccessTokenMapOutput) ToTeamAccessTokenMapOutputWithContext(ctx context.Context) TeamAccessTokenMapOutput {
	return o
}

func (o TeamAccessTokenMapOutput) MapIndex(k pulumi.StringInput) TeamAccessTokenOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *TeamAccessToken {
		return vs[0].(map[string]*TeamAccessToken)[vs[1].(string)]
	}).(TeamAccessTokenOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*TeamAccessTokenInput)(nil)).Elem(), &TeamAccessToken{})
	pulumi.RegisterInputType(reflect.TypeOf((*TeamAccessTokenArrayInput)(nil)).Elem(), TeamAccessTokenArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*TeamAccessTokenMapInput)(nil)).Elem(), TeamAccessTokenMap{})
	pulumi.RegisterOutputType(TeamAccessTokenOutput{})
	pulumi.RegisterOutputType(TeamAccessTokenArrayOutput{})
	pulumi.RegisterOutputType(TeamAccessTokenMapOutput{})
}
