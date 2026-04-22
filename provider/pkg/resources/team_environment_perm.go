package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

type PulumiServiceTeamEnvironmentPermissionResource struct {
	Client pulumiapi.TeamClient
}

type TeamEnvironmentPermissionInput struct {
	Organization    string `pulumi:"organization"`
	Team            string `pulumi:"team"`
	Environment     string `pulumi:"environment"`
	Project         string `pulumi:"project"`
	Permission      string `pulumi:"permission"`
	MaxOpenDuration string `pulumi:"maxOpenDuration"`
}

func (i *TeamEnvironmentPermissionInput) ToPropertyMap() resource.PropertyMap {
	return util.ToPropertyMap(*i)
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) ToPulumiServiceTeamInput(
	inputMap resource.PropertyMap,
) (*TeamEnvironmentPermissionInput, error) {
	input := TeamEnvironmentPermissionInput{}
	return &input, util.FromPropertyMap(inputMap, &input)
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Name() string {
	return "pulumiservice:index:TeamEnvironmentPermission"
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Check(
	req *pulumirpc.CheckRequest,
) (*pulumirpc.CheckResponse, error) {
	// Work on the property map directly so we only touch fields the user
	// actually supplied. Re-serializing the input struct would emit every
	// tagged field (including empty strings), which introduces spurious
	// diffs for optional fields that weren't present in state written by
	// older provider versions.
	news, err := plugin.UnmarshalProperties(req.GetNews(), util.StandardUnmarshal)
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure
	if prop, ok := news["maxOpenDuration"]; ok {
		if !prop.IsString() {
			// Reject non-string values deterministically at preview; otherwise
			// util.FromPropertyMap panics during Create when asserting a
			// number into the string field.
			failures = append(failures, &pulumirpc.CheckFailure{
				Property: "maxOpenDuration",
				Reason:   "maxOpenDuration property is present but can't be parsed as string",
			})
		} else {
			raw := prop.StringValue()
			if raw == "" {
				// Treat an explicit empty string as "unset" to stay consistent
				// with state saved before this field existed.
				delete(news, "maxOpenDuration")
			} else {
				d, err := time.ParseDuration(raw)
				if err != nil {
					failures = append(failures, &pulumirpc.CheckFailure{
						Property: "maxOpenDuration",
						Reason:   fmt.Sprintf("malformed duration: %v", err),
					})
				} else if normalized := d.String(); normalized != raw {
					// Normalize the duration to prevent spurious diffs.
					news["maxOpenDuration"] = resource.NewStringProperty(normalized)
				}
			}
		}
	}

	inputs, err := plugin.MarshalProperties(news, util.StandardMarshal)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CheckResponse{
		Inputs:   inputs,
		Failures: failures,
	}, nil
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Read(
	req *pulumirpc.ReadRequest,
) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()
	permID, err := splitTeamEnvironmentPermissionID(req.GetId())
	if err != nil {
		return nil, err
	}

	request := pulumiapi.TeamEnvironmentSettingsRequest{
		Organization: permID.Organization,
		Team:         permID.Team,
		Environment:  permID.Environment,
		Project:      permID.Project,
	}
	permission, maxOpenDuration, err := tp.Client.GetTeamEnvironmentSettings(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get team environment permission: %w", err)
	}
	if permission == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	inputs := TeamEnvironmentPermissionInput{
		Organization: permID.Organization,
		Team:         permID.Team,
		Project:      permID.Project,
		Environment:  permID.Environment,
		Permission:   *permission,
	}
	if maxOpenDuration != nil {
		inputs.MaxOpenDuration = (time.Duration)(*maxOpenDuration).String()
	}

	// Omit maxOpenDuration when it wasn't set on the remote resource; emitting
	// an empty string would cause a spurious replacement on the next update
	// against state written by providers that didn't have this field.
	propertyMap := inputs.ToPropertyMap()
	if inputs.MaxOpenDuration == "" {
		delete(propertyMap, "maxOpenDuration")
	}

	properties, err := plugin.MarshalProperties(propertyMap, plugin.MarshalOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inputs to properties: %w", err)
	}
	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: properties,
		Inputs:     properties,
	}, nil
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Create(
	req *pulumirpc.CreateRequest,
) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	var input TeamEnvironmentPermissionInput
	err := util.FromProperties(req.GetProperties(), &input)
	if err != nil {
		return nil, err
	}

	var maxOpenDuration *pulumiapi.Duration
	if input.MaxOpenDuration != "" {
		d, err := time.ParseDuration(input.MaxOpenDuration)
		if err != nil {
			return nil, err
		}
		maxOpenDuration = (*pulumiapi.Duration)(&d)
	}

	request := pulumiapi.CreateTeamEnvironmentSettingsRequest{
		TeamEnvironmentSettingsRequest: pulumiapi.TeamEnvironmentSettingsRequest{
			Organization: input.Organization,
			Team:         input.Team,
			Project:      input.Project,
			Environment:  input.Environment,
		},
		Permission:      input.Permission,
		MaxOpenDuration: maxOpenDuration,
	}

	err = tp.Client.AddEnvironmentSettings(ctx, request)
	if err != nil {
		return nil, err
	}

	environmentPermissionID := teamEnvironmentPermissionID{
		Organization: input.Organization,
		Team:         input.Team,
		Project:      input.Project,
		Environment:  input.Environment,
	}

	return &pulumirpc.CreateResponse{
		Id:         environmentPermissionID.String(),
		Properties: req.GetProperties(),
	}, nil
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	var input TeamEnvironmentPermissionInput
	err := util.FromProperties(req.GetProperties(), &input)
	if err != nil {
		return nil, err
	}
	request := pulumiapi.TeamEnvironmentSettingsRequest{
		Organization: input.Organization,
		Team:         input.Team,
		Project:      input.Project,
		Environment:  input.Environment,
	}
	err = tp.Client.RemoveEnvironmentSettings(ctx, request)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Diff(
	req *pulumirpc.DiffRequest,
) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(
		req.GetOldInputs(),
		plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}
	news, err := plugin.UnmarshalProperties(
		req.GetNews(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: false},
	)
	if err != nil {
		return nil, err
	}

	// Provider versions 0.29.3–0.36.0 wrote `maxOpenDuration: ""` into state
	// whenever the user did not set the field. #752 fixed the Check/Read
	// paths, but state saved by those versions still carries the empty
	// string. Without this normalization, the first preview after upgrading
	// from that window would observe the key as deleted and force one more
	// spurious replacement — the exact failure mode #751 set out to remove.
	normalizeEmptyMaxOpenDuration(olds)
	normalizeEmptyMaxOpenDuration(news)

	var changedKeys []string
	for _, k := range olds.Diff(news).ChangedKeys() {
		changedKeys = append(changedKeys, string(k))
	}

	changes := pulumirpc.DiffResponse_DIFF_NONE
	if len(changedKeys) > 0 {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}
	return &pulumirpc.DiffResponse{
		Changes:             changes,
		Replaces:            changedKeys,
		DeleteBeforeReplace: true,
	}, nil
}

// normalizeEmptyMaxOpenDuration removes a zero-value `maxOpenDuration` so it
// compares equal to an absent field. See Diff() for the history.
func normalizeEmptyMaxOpenDuration(m resource.PropertyMap) {
	if v, ok := m["maxOpenDuration"]; ok && v.IsString() && v.StringValue() == "" {
		delete(m, "maxOpenDuration")
	}
}

// Update does nothing because we always replace on changes, never an update
func (tp *PulumiServiceTeamEnvironmentPermissionResource) Update(
	_ *pulumirpc.UpdateRequest,
) (*pulumirpc.UpdateResponse, error) {
	return nil, fmt.Errorf("unexpected call to update, expected create to be called instead")
}

type teamEnvironmentPermissionID struct {
	Organization string
	Team         string
	Project      string
	Environment  string
}

func (s *teamEnvironmentPermissionID) String() string {
	return fmt.Sprintf("%s/%s/%s+%s", s.Organization, s.Team, s.Project, s.Environment)
}

func splitTeamEnvironmentPermissionID(id string) (teamEnvironmentPermissionID, error) {
	split := strings.Split(id, "/")
	if len(split) != 3 {
		return teamEnvironmentPermissionID{}, fmt.Errorf("invalid id %q, expected 3 parts", id)
	}

	splitProjectEnv := strings.Split(split[2], "+")
	if len(splitProjectEnv) == 1 {
		return teamEnvironmentPermissionID{
			Organization: split[0],
			Team:         split[1],
			Project:      "default",
			Environment:  splitProjectEnv[0],
		}, nil
	}
	if len(splitProjectEnv) == 2 {
		return teamEnvironmentPermissionID{
			Organization: split[0],
			Team:         split[1],
			Project:      splitProjectEnv[0],
			Environment:  splitProjectEnv[1],
		}, nil
	}

	return teamEnvironmentPermissionID{}, fmt.Errorf(
		"invalid id %q, expected environment name or project/environment in last part",
		id,
	)
}
