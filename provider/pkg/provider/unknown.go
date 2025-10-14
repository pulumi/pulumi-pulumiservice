package provider

import (
	"fmt"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

// PulumiServiceUnknownResource represents an unknown resource type.
type PulumiServiceUnknownResource struct{}

// PulumiServiceUnknownFunction represents an unknown function.
type PulumiServiceUnknownFunction struct{}

// Name returns the name of the unknown resource.
func (u *PulumiServiceUnknownResource) Name() string {
	return "pulumiservice:index:Unknown"
}

// Configure configures the unknown resource.
func (u *PulumiServiceUnknownResource) Configure(_ PulumiServiceConfig) {
}

// Diff returns an error for unknown resources.
func (u *PulumiServiceUnknownResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

// Delete returns an error for unknown resources.
func (u *PulumiServiceUnknownResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

// Create returns an error for unknown resources.
func (u *PulumiServiceUnknownResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

// Check returns an error for unknown resources.
func (u *PulumiServiceUnknownResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

// Update returns an error for unknown resources.
func (u *PulumiServiceUnknownResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

// Read returns an error for unknown resources.
func (u *PulumiServiceUnknownResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func createUnknownResourceErrorFromRequest(req ResourceBase) error {
	rn := getResourceNameFromRequest(req)
	return fmt.Errorf("unknown resource type '%s'", rn)
}

// Invoke returns an error for unknown functions.
func (u *PulumiServiceUnknownResource) Invoke(
	_ *pulumiserviceProvider, req *pulumirpc.InvokeRequest,
) (*pulumirpc.InvokeResponse, error) {
	return &pulumirpc.InvokeResponse{Return: nil}, fmt.Errorf("unknown function '%s'", req.Tok)
}

// Name returns the name of the unknown function.
func (f *PulumiServiceUnknownFunction) Name() string {
	return "pulumiservice:index:Unknown"
}

// Configure configures the unknown function.
func (f *PulumiServiceUnknownFunction) Configure(_ PulumiServiceConfig) {
}
