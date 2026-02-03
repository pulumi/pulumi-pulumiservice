package provider

import (
	"fmt"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceUnknownResource struct{}
type PulumiServiceUnknownFunction struct{}

func (u *PulumiServiceUnknownResource) Name() string {
	return "pulumiservice:index:Unknown"
}

func (u *PulumiServiceUnknownResource) Configure(_ PulumiServiceConfig) {
}

func (u *PulumiServiceUnknownResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func (u *PulumiServiceUnknownResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func (u *PulumiServiceUnknownResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func (u *PulumiServiceUnknownResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func (u *PulumiServiceUnknownResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func (u *PulumiServiceUnknownResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func createUnknownResourceErrorFromRequest(req ResourceBase) error {
	rn := getResourceNameFromRequest(req)
	return fmt.Errorf("unknown resource type '%s'", rn)
}

func (u *PulumiServiceUnknownResource) Invoke(
	_ *pulumiserviceProvider,
	req *pulumirpc.InvokeRequest,
) (*pulumirpc.InvokeResponse, error) {
	return &pulumirpc.InvokeResponse{Return: nil}, fmt.Errorf("unknown function '%s'", req.Tok)
}

func (f *PulumiServiceUnknownFunction) Name() string {
	return "pulumiservice:index:Unknown"
}

func (f *PulumiServiceUnknownFunction) Configure(_ PulumiServiceConfig) {
}
