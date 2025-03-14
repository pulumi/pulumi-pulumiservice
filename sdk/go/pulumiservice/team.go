// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package pulumiservice

import (
	"context"
	"reflect"

	"errors"
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The Pulumi Cloud offers role-based access control (RBAC) using teams. Teams allow organization admins to assign a set of stack permissions to a group of users.
type Team struct {
	pulumi.CustomResourceState

	// Optional. Team description.
	Description pulumi.StringPtrOutput `pulumi:"description"`
	// Optional. Team display name.
	DisplayName pulumi.StringPtrOutput `pulumi:"displayName"`
	// The GitHub ID of the team to mirror. Must be in the same GitHub organization that the Pulumi org is backed by. Required for "github" teams.
	GithubTeamId pulumi.Float64PtrOutput `pulumi:"githubTeamId"`
	// List of Pulumi Cloud usernames of team members.
	Members pulumi.StringArrayOutput `pulumi:"members"`
	// The team's name. Required for "pulumi" teams.
	Name pulumi.StringPtrOutput `pulumi:"name"`
	// The name of the Pulumi organization the team belongs to.
	OrganizationName pulumi.StringOutput `pulumi:"organizationName"`
	// The type of team. Must be either `pulumi` or `github`.
	TeamType pulumi.StringOutput `pulumi:"teamType"`
}

// NewTeam registers a new resource with the given unique name, arguments, and options.
func NewTeam(ctx *pulumi.Context,
	name string, args *TeamArgs, opts ...pulumi.ResourceOption) (*Team, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.OrganizationName == nil {
		return nil, errors.New("invalid value for required argument 'OrganizationName'")
	}
	if args.TeamType == nil {
		return nil, errors.New("invalid value for required argument 'TeamType'")
	}
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource Team
	err := ctx.RegisterResource("pulumiservice:index:Team", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetTeam gets an existing Team resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetTeam(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *TeamState, opts ...pulumi.ResourceOption) (*Team, error) {
	var resource Team
	err := ctx.ReadResource("pulumiservice:index:Team", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering Team resources.
type teamState struct {
}

type TeamState struct {
}

func (TeamState) ElementType() reflect.Type {
	return reflect.TypeOf((*teamState)(nil)).Elem()
}

type teamArgs struct {
	// Optional. Team description.
	Description *string `pulumi:"description"`
	// Optional. Team display name.
	DisplayName *string `pulumi:"displayName"`
	// The GitHub ID of the team to mirror. Must be in the same GitHub organization that the Pulumi org is backed by. Required for "github" teams.
	GithubTeamId *float64 `pulumi:"githubTeamId"`
	// List of Pulumi Cloud usernames of team members.
	Members []string `pulumi:"members"`
	// The team's name. Required for "pulumi" teams.
	Name *string `pulumi:"name"`
	// The name of the Pulumi organization the team belongs to.
	OrganizationName string `pulumi:"organizationName"`
	// The type of team. Must be either `pulumi` or `github`.
	TeamType string `pulumi:"teamType"`
}

// The set of arguments for constructing a Team resource.
type TeamArgs struct {
	// Optional. Team description.
	Description pulumi.StringPtrInput
	// Optional. Team display name.
	DisplayName pulumi.StringPtrInput
	// The GitHub ID of the team to mirror. Must be in the same GitHub organization that the Pulumi org is backed by. Required for "github" teams.
	GithubTeamId pulumi.Float64PtrInput
	// List of Pulumi Cloud usernames of team members.
	Members pulumi.StringArrayInput
	// The team's name. Required for "pulumi" teams.
	Name pulumi.StringPtrInput
	// The name of the Pulumi organization the team belongs to.
	OrganizationName pulumi.StringInput
	// The type of team. Must be either `pulumi` or `github`.
	TeamType pulumi.StringInput
}

func (TeamArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*teamArgs)(nil)).Elem()
}

type TeamInput interface {
	pulumi.Input

	ToTeamOutput() TeamOutput
	ToTeamOutputWithContext(ctx context.Context) TeamOutput
}

func (*Team) ElementType() reflect.Type {
	return reflect.TypeOf((**Team)(nil)).Elem()
}

func (i *Team) ToTeamOutput() TeamOutput {
	return i.ToTeamOutputWithContext(context.Background())
}

func (i *Team) ToTeamOutputWithContext(ctx context.Context) TeamOutput {
	return pulumi.ToOutputWithContext(ctx, i).(TeamOutput)
}

// TeamArrayInput is an input type that accepts TeamArray and TeamArrayOutput values.
// You can construct a concrete instance of `TeamArrayInput` via:
//
//	TeamArray{ TeamArgs{...} }
type TeamArrayInput interface {
	pulumi.Input

	ToTeamArrayOutput() TeamArrayOutput
	ToTeamArrayOutputWithContext(context.Context) TeamArrayOutput
}

type TeamArray []TeamInput

func (TeamArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*Team)(nil)).Elem()
}

func (i TeamArray) ToTeamArrayOutput() TeamArrayOutput {
	return i.ToTeamArrayOutputWithContext(context.Background())
}

func (i TeamArray) ToTeamArrayOutputWithContext(ctx context.Context) TeamArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(TeamArrayOutput)
}

// TeamMapInput is an input type that accepts TeamMap and TeamMapOutput values.
// You can construct a concrete instance of `TeamMapInput` via:
//
//	TeamMap{ "key": TeamArgs{...} }
type TeamMapInput interface {
	pulumi.Input

	ToTeamMapOutput() TeamMapOutput
	ToTeamMapOutputWithContext(context.Context) TeamMapOutput
}

type TeamMap map[string]TeamInput

func (TeamMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*Team)(nil)).Elem()
}

func (i TeamMap) ToTeamMapOutput() TeamMapOutput {
	return i.ToTeamMapOutputWithContext(context.Background())
}

func (i TeamMap) ToTeamMapOutputWithContext(ctx context.Context) TeamMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(TeamMapOutput)
}

type TeamOutput struct{ *pulumi.OutputState }

func (TeamOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**Team)(nil)).Elem()
}

func (o TeamOutput) ToTeamOutput() TeamOutput {
	return o
}

func (o TeamOutput) ToTeamOutputWithContext(ctx context.Context) TeamOutput {
	return o
}

// Optional. Team description.
func (o TeamOutput) Description() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *Team) pulumi.StringPtrOutput { return v.Description }).(pulumi.StringPtrOutput)
}

// Optional. Team display name.
func (o TeamOutput) DisplayName() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *Team) pulumi.StringPtrOutput { return v.DisplayName }).(pulumi.StringPtrOutput)
}

// The GitHub ID of the team to mirror. Must be in the same GitHub organization that the Pulumi org is backed by. Required for "github" teams.
func (o TeamOutput) GithubTeamId() pulumi.Float64PtrOutput {
	return o.ApplyT(func(v *Team) pulumi.Float64PtrOutput { return v.GithubTeamId }).(pulumi.Float64PtrOutput)
}

// List of Pulumi Cloud usernames of team members.
func (o TeamOutput) Members() pulumi.StringArrayOutput {
	return o.ApplyT(func(v *Team) pulumi.StringArrayOutput { return v.Members }).(pulumi.StringArrayOutput)
}

// The team's name. Required for "pulumi" teams.
func (o TeamOutput) Name() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *Team) pulumi.StringPtrOutput { return v.Name }).(pulumi.StringPtrOutput)
}

// The name of the Pulumi organization the team belongs to.
func (o TeamOutput) OrganizationName() pulumi.StringOutput {
	return o.ApplyT(func(v *Team) pulumi.StringOutput { return v.OrganizationName }).(pulumi.StringOutput)
}

// The type of team. Must be either `pulumi` or `github`.
func (o TeamOutput) TeamType() pulumi.StringOutput {
	return o.ApplyT(func(v *Team) pulumi.StringOutput { return v.TeamType }).(pulumi.StringOutput)
}

type TeamArrayOutput struct{ *pulumi.OutputState }

func (TeamArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*Team)(nil)).Elem()
}

func (o TeamArrayOutput) ToTeamArrayOutput() TeamArrayOutput {
	return o
}

func (o TeamArrayOutput) ToTeamArrayOutputWithContext(ctx context.Context) TeamArrayOutput {
	return o
}

func (o TeamArrayOutput) Index(i pulumi.IntInput) TeamOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *Team {
		return vs[0].([]*Team)[vs[1].(int)]
	}).(TeamOutput)
}

type TeamMapOutput struct{ *pulumi.OutputState }

func (TeamMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*Team)(nil)).Elem()
}

func (o TeamMapOutput) ToTeamMapOutput() TeamMapOutput {
	return o
}

func (o TeamMapOutput) ToTeamMapOutputWithContext(ctx context.Context) TeamMapOutput {
	return o
}

func (o TeamMapOutput) MapIndex(k pulumi.StringInput) TeamOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *Team {
		return vs[0].(map[string]*Team)[vs[1].(string)]
	}).(TeamOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*TeamInput)(nil)).Elem(), &Team{})
	pulumi.RegisterInputType(reflect.TypeOf((*TeamArrayInput)(nil)).Elem(), TeamArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*TeamMapInput)(nil)).Elem(), TeamMap{})
	pulumi.RegisterOutputType(TeamOutput{})
	pulumi.RegisterOutputType(TeamArrayOutput{})
	pulumi.RegisterOutputType(TeamMapOutput{})
}
