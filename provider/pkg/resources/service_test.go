package resources

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
)

type getServiceFunc func() (*pulumiapi.Service, error)

type ServiceClientMock struct {
	getServiceFunc getServiceFunc
}

func (c *ServiceClientMock) GetService(ctx context.Context, orgName, ownerType, ownerName, serviceName string) (*pulumiapi.Service, error) {
	return c.getServiceFunc()
}

func (c *ServiceClientMock) CreateService(ctx context.Context, req pulumiapi.CreateServiceRequest) (*pulumiapi.Service, error) {
	return nil, nil
}

func (c *ServiceClientMock) UpdateService(ctx context.Context, req pulumiapi.UpdateServiceRequest) (*pulumiapi.Service, error) {
	return nil, nil
}

func (c *ServiceClientMock) DeleteService(ctx context.Context, orgName, ownerType, ownerName, serviceName string, force bool) error {
	return nil
}

func (c *ServiceClientMock) AddServiceItem(ctx context.Context, req pulumiapi.AddServiceItemRequest) error {
	return nil
}

func (c *ServiceClientMock) RemoveServiceItem(ctx context.Context, req pulumiapi.RemoveServiceItemRequest) error {
	return nil
}

func buildServiceClientMock(getServiceFunc getServiceFunc) *ServiceClientMock {
	return &ServiceClientMock{
		getServiceFunc: getServiceFunc,
	}
}

func TestService(t *testing.T) {
	t.Run("Read when the resource is not found", func(t *testing.T) {
		mockedClient := buildServiceClientMock(
			func() (*pulumiapi.Service, error) { return nil, nil },
		)

		provider := PulumiServiceServiceResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "myorg/user/myuser/myservice",
			Urn: "urn:123",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, "", resp.Id)
		assert.Nil(t, resp.Properties)
	})

	t.Run("Read when the resource is found", func(t *testing.T) {
		mockedClient := buildServiceClientMock(
			func() (*pulumiapi.Service, error) {
				return &pulumiapi.Service{
					Name:        "myservice",
					Description: "My test service",
					OwnerType:   "user",
					OwnerName:   "myuser",
					Properties: map[string]string{
						"key1": "value1",
					},
					Items: []pulumiapi.ServiceItem{
						{
							ItemType: "stack",
							Name:     "myorg/myproject/mystack",
						},
					},
				}, nil
			},
		)

		provider := PulumiServiceServiceResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "myorg/user/myuser/myservice",
			Urn: "urn:123",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, "myorg/user/myuser/myservice", resp.Id)
		assert.NotNil(t, resp.Properties)
	})

	t.Run("SplitServiceID with valid ID", func(t *testing.T) {
		orgName, ownerType, ownerName, serviceName, err := splitServiceID("myorg/user/myuser/myservice")

		assert.NoError(t, err)
		assert.Equal(t, "myorg", orgName)
		assert.Equal(t, "user", ownerType)
		assert.Equal(t, "myuser", ownerName)
		assert.Equal(t, "myservice", serviceName)
	})

	t.Run("SplitServiceID with invalid ID", func(t *testing.T) {
		_, _, _, _, err := splitServiceID("invalid/id")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a valid service ID")
	})

	t.Run("GenerateServiceID", func(t *testing.T) {
		id := generateServiceID("myorg", "team", "myteam", "myservice")

		assert.Equal(t, "myorg/team/myteam/myservice", id)
	})
}
