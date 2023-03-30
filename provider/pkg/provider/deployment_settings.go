package provider

import (
	"context"
	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/serde"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/ryboe/q"
	"path"
)

type PulumiServiceDeploymentSettingsInput struct {
	Organization     string                `pulumi:"organization"`
	Project          string                `pulumi:"project"`
	Stack            string                `pulumi:"stack"`
	SourceContext    SourceContextInput    `pulumi:"sourceContext"`
	ExecutorContext  ExecutorContextInput  `pulumi:"executorContext"`
	OperationContext OperationContextInput `pulumi:"operationContext"`
	GitHub           GitHubInput           `pulumi:"github"`
}

type SourceContextInput struct {
	Git GitSourceInput `pulumi:"git"`
}

type GitSourceInput struct {
	RepoURL string       `pulumi:"repoUrl"`
	Branch  string       `pulumi:"branch"`
	Commit  string       `pulumi:"commit"`
	RepoDir string       `pulumi:"repoDir"`
	GitAuth GitAuthInput `pulumi:"gitAuth"`
}

type GitAuthInput struct {
	PersonalAccessToken string                `pulumi:"personalAccessToken"`
	SSHAuth             GitAuthSSHAuthInput   `pulumi:"sshAuth"`
	BasicAuth           GitAuthBasicAuthInput `pulumi:"basicAuth"`
}

type GitAuthSSHAuthInput struct {
	SSHPrivateKey string `pulumi:"sshPrivateKey"`
	Password      string `pulumi:"password"`
}

type GitAuthBasicAuthInput struct {
	Username string `pulumi:"username"`
	Password string `pulumi:"password"`
}

type ExecutorContextInput struct {
	ExecutorImage string `pulumi:"executorImage"`
}

type OperationContextInput struct {
	Options              OperationContextOptionsInput `pulumi:"options"`
	PreRunCommands       []string                     `pulumi:"preRunCommands"`
	EnvironmentVariables map[string]string            `pulumi:"environmentVariables"`
	OIDC                 OperationContextOIDCInput    `pulumi:"oidc"`
}

type OperationContextOIDCInput struct {
	AWS   AWSOIDCInput   `pulumi:"aws"`
	GCP   GCPOIDCInput   `pulumi:"gcp"`
	Azure AzureOIDCInput `pulumi:"azure"`
}

type AWSOIDCInput struct {
	Duration    int      `pulumi:"duration"`
	PolicyARNs  []string `pulumi:"policyARNs"`
	RoleARN     string   `pulumi:"roleARN"`
	SessionName string   `pulumi:"sessionName"`
}

type GCPOIDCInput struct {
	ProjectID      string `pulumi:"projectId"`
	Region         string `pulumi:"region"`
	WorkloadPoolID string `pulumi:"workloadPoolId"`
	ProviderID     string `pulumi:"providerId"`
	ServiceAccount string `pulumi:"serviceAccount"`
	TokenLifetime  int    `pulumi:"tokenLifetime"`
}

type AzureOIDCInput struct {
	ClientID       string `pulumi:"clientId"`
	TenantID       string `pulumi:"tenantId"`
	SubscriptionID string `pulumi:"subscriptionId"`
}

type OperationContextOptionsInput struct {
	SkipInstallDependencies bool   `pulumi:"skipInstallDependencies"`
	Shell                   string `pulumi:"shell"`
}

type GitHubInput struct {
	Repository          string   `pulumi:"repository"`
	DeployCommits       bool     `pulumi:"deployCommits"`
	PreviewPullRequests bool     `pulumi:"previewPullRequests"`
	Paths               []string `pulumi:"paths"`
}

func (i *PulumiServiceDeploymentSettingsInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["organization"] = resource.NewPropertyValue(i.Organization)
	pm["project"] = resource.NewPropertyValue(i.Project)
	pm["stack"] = resource.NewPropertyValue(i.Stack)
	return serde.ToPropertyMap(*i, structTagKey)
}

type PulumiServiceDeploymentSettingsResource struct {
	client *pulumiapi.Client
}

func (ds *PulumiServiceDeploymentSettingsResource) ToPulumiServiceDeploymentSettingsInput(inputMap resource.PropertyMap) PulumiServiceDeploymentSettingsInput {
	input := PulumiServiceDeploymentSettingsInput{}

	if inputMap["organization"].HasValue() && inputMap["organization"].IsString() {
		input.Organization = inputMap["organization"].StringValue()
	}

	if inputMap["description"].HasValue() && inputMap["description"].IsString() {
		input.Project = inputMap["project"].StringValue()
	}

	if inputMap["organizationName"].HasValue() && inputMap["organizationName"].IsString() {
		input.Stack = inputMap["stack"].StringValue()
	}

	return input
}

func (ds *PulumiServiceDeploymentSettingsResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}
	q.Q(olds, news)

	diffs := olds["__inputs"].ObjectValue().Diff(news)
	q.Q(diffs)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes: pulumirpc.DiffResponse_DIFF_NONE,
		}, nil
	}

	changes := pulumirpc.DiffResponse_DIFF_NONE
	if diffs.AnyChanges() {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}

	return &pulumirpc.DiffResponse{
		Changes: changes,
	}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Configure(config PulumiServiceConfig) {}

func (ds *PulumiServiceDeploymentSettingsResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return &pulumirpc.ReadResponse{}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	var inputs PulumiServiceDeploymentSettingsInput
	err := serde.FromProperties(req.GetProperties(), structTagKey, &inputs)
	if err != nil {
		return nil, err
	}
	stackName := pulumiapi.StackName{
		OrgName:     inputs.Organization,
		ProjectName: inputs.Project,
		StackName:   inputs.Stack,
	}
	err = ds.client.DeleteDeploymentSettings(ctx, stackName)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	q.Q(req.GetProperties())
	inputsMap, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}
	q.Q(inputsMap)
	inputs := ds.ToPulumiServiceDeploymentSettingsInput(inputsMap)
	stack := pulumiapi.StackName{
		OrgName:     inputs.Organization,
		ProjectName: inputs.Project,
		StackName:   inputs.Stack,
	}
	settings := pulumiapi.DeploymentSettings{
		OperationContext: &pulumiapi.OperationContext{
			PreRunCommands: []string{"yarn"},
		},
	}
	err = ds.client.CreateDeploymentSettings(ctx, stack, settings)
	if err != nil {
		return nil, err
	}
	q.Q(req.GetProperties(), inputs)
	return &pulumirpc.CreateResponse{
		Id:         path.Join(stack.OrgName, stack.ProjectName, stack.StackName),
		Properties: req.GetProperties(),
	}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return &pulumirpc.UpdateResponse{}, nil
}

func (ds *PulumiServiceDeploymentSettingsResource) Name() string {
	return "pulumiservice:index:DeploymentSettings"
}
