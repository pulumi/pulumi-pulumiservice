package provider

import (
	"fmt"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	rpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceUnknownResource struct{}
type PulumiServiceUnknownFunction struct{}

func (c PulumiServiceUnknownResource) Name() string {
	return "pulumi-service:index:Unknown"
}

func (u *PulumiServiceUnknownResource) Configure(config PulumiServiceConfig) {
}

func (c *PulumiServiceUnknownResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func (c *PulumiServiceUnknownResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func (c *PulumiServiceUnknownResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func (k *PulumiServiceUnknownResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func (k *PulumiServiceUnknownResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func (k *PulumiServiceUnknownResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func createUnknownResourceErrorFromRequest(req ResourceBase) error {
	rn := getResourceNameFromRequest(req)
	return fmt.Errorf("unknown resource type '%s'", rn)
}

func (f *PulumiServiceUnknownResource) Invoke(s *pulumiserviceProvider, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	return &rpc.InvokeResponse{Return: nil}, fmt.Errorf("unknown function '%s'", req.Tok)
}

func (f *PulumiServiceUnknownFunction) Name() string {
	return "pulumi-service:index:Unknown"
}

func (f *PulumiServiceUnknownFunction) Configure(config PulumiServiceConfig) {
}
