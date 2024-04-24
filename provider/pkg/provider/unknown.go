package provider

import (
	"context"
	"fmt"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceUnknownResource struct{}
type PulumiServiceUnknownFunction struct{}

func (u *PulumiServiceUnknownResource) Name() string {
	return "pulumiservice:index:Unknown"
}

func (u *PulumiServiceUnknownResource) Diff(_ context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func (u *PulumiServiceUnknownResource) Delete(_ context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func (u *PulumiServiceUnknownResource) Create(_ context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func (u *PulumiServiceUnknownResource) Check(_ context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func (u *PulumiServiceUnknownResource) Update(_ context.Context, req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func (u *PulumiServiceUnknownResource) Read(_ context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func createUnknownResourceErrorFromRequest(req ResourceBase) error {
	rn := getResourceNameFromRequest(req)
	return fmt.Errorf("unknown resource type '%s'", rn)
}

func (u *PulumiServiceUnknownResource) Invoke(_ context.Context, s *pulumiserviceProvider, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	return &pulumirpc.InvokeResponse{Return: nil}, fmt.Errorf("unknown function '%s'", req.Tok)
}

func (f *PulumiServiceUnknownFunction) Name() string {
	return "pulumiservice:index:Unknown"
}
