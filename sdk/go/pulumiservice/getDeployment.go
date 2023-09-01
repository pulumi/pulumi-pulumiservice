// Code generated by pulumigen DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package pulumiservice

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Gets information about a deployment.
func GetDeployment(ctx *pulumi.Context, args *GetDeploymentArgs, opts ...pulumi.InvokeOption) (*GetDeploymentResult, error) {
	var rv GetDeploymentResult
	err := ctx.Invoke("pulumiservice:index:GetDeployment", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

type GetDeploymentArgs struct {
	DeploymentId string `pulumi:"deploymentId"`
	Organization string `pulumi:"organization"`
	Project      string `pulumi:"project"`
	Stack        string `pulumi:"stack"`
}

type GetDeploymentResult struct {
	Status *string `pulumi:"status"`
}

func GetDeploymentOutput(ctx *pulumi.Context, args GetDeploymentOutputArgs, opts ...pulumi.InvokeOption) GetDeploymentResultOutput {
	return pulumi.ToOutputWithContext(context.Background(), args).
		ApplyT(func(v interface{}) (GetDeploymentResult, error) {
			args := v.(GetDeploymentArgs)
			r, err := GetDeployment(ctx, &args, opts...)
			var s GetDeploymentResult
			if r != nil {
				s = *r
			}
			return s, err
		}).(GetDeploymentResultOutput)
}

type GetDeploymentOutputArgs struct {
	DeploymentId pulumi.StringInput `pulumi:"deploymentId"`
	Organization pulumi.StringInput `pulumi:"organization"`
	Project      pulumi.StringInput `pulumi:"project"`
	Stack        pulumi.StringInput `pulumi:"stack"`
}

func (GetDeploymentOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*GetDeploymentArgs)(nil)).Elem()
}

type GetDeploymentResultOutput struct{ *pulumi.OutputState }

func (GetDeploymentResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*GetDeploymentResult)(nil)).Elem()
}

func (o GetDeploymentResultOutput) ToGetDeploymentResultOutput() GetDeploymentResultOutput {
	return o
}

func (o GetDeploymentResultOutput) ToGetDeploymentResultOutputWithContext(ctx context.Context) GetDeploymentResultOutput {
	return o
}

func (o GetDeploymentResultOutput) Status() pulumi.StringPtrOutput {
	return o.ApplyT(func(v GetDeploymentResult) *string { return v.Status }).(pulumi.StringPtrOutput)
}

func init() {
	pulumi.RegisterOutputType(GetDeploymentResultOutput{})
}
