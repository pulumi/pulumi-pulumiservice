package provider

import (
	"fmt"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"golang.org/x/exp/slices"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

var (
	// Life-cycle participation
	_ infer.CustomResource[TeamInput, TeamState] = (*Team)(nil)
	_ infer.CustomCheck[TeamInput]               = (*Team)(nil)
	_ infer.CustomDelete[TeamState]              = (*Team)(nil)
	_ infer.CustomRead[TeamInput, TeamState]     = (*Team)(nil)
	_ infer.CustomUpdate[TeamInput, TeamState]   = (*Team)(nil)

	// Schema documentation
	_ infer.Annotated = (*Team)(nil)
	_ infer.Annotated = (*TeamInput)(nil)
	_ infer.Annotated = (*TeamState)(nil)
)

type Team struct{}

func (p *Team) Annotate(a infer.Annotator) {
	a.Describe(p, "The Pulumi Cloud offers role-based access control (RBAC) using teams. "+
		"Teams allow organization admins to assign a set of stack permissions to a group of users.")
}

type TeamInput struct {
	TeamType         string   `pulumi:"teamType"` // TODO[BREAKING]: This should be an enum
	Name             string   `pulumi:"name,optional"`
	DisplayName      string   `pulumi:"displayName,optional"`
	Description      string   `pulumi:"description,optional"`
	Members          []string `pulumi:"members,optional"`
	OrganizationName string   `pulumi:"organizationName"`
	GithubTeamId     int64    `pulumi:"githubTeamId,optional"`
}

func (p *TeamInput) Annotate(a infer.Annotator) {
	a.Describe(&p.TeamType, "The type of team. Must be either `pulumi` or `github`.")
	a.Describe(&p.Name, "The team's name. Required for \"pulumi\" teams.")
	a.Describe(&p.DisplayName, "Optional. Team display name.")
	a.Describe(&p.Description, "Optional. Team description.")
	a.Describe(&p.Members, "List of team members.")
	a.Describe(&p.OrganizationName, "The name of the Pulumi organization the team belongs to.")
	a.Describe(&p.GithubTeamId, "The GitHub ID of the team to mirror. Must be in the same GitHub "+
		"organization that the Pulumi org is backed by. Required for \"github\" teams.")
}

type TeamState struct {
	TeamInput
	Members []string `pulumi:"members"`
}

func (p *TeamState) Annotate(a infer.Annotator) {
	// TODO: It's not clear if this is necessary. p.Members is a required where
	// p.Input.Members is optional, but it might be ok to omit the description here.
	a.Describe(&p.Members, "List of team members.")
}

func (*Team) Check(
	ctx p.Context, name string, oldInputs resource.PropertyMap, newInputs resource.PropertyMap,
) (TeamInput, []p.CheckFailure, error) {
	inputs, failures, err := infer.DefaultCheck[TeamInput](newInputs)
	if len(failures) > 0 || err != nil {
		return inputs, failures, err
	}

	switch inputs.TeamType {
	case "github":
		if inputs.GithubTeamId == 0 {
			failures = append(failures, p.CheckFailure{
				Reason:   "teams with teamType 'github' require a 'githubTeamId'",
				Property: "githubTeamId",
			})
		}
	case "pulumi":
		if len(inputs.Name) == 0 {
			failures = append(failures, p.CheckFailure{
				Reason:   "teams with teamType 'pulumi' require a 'name'",
				Property: "name",
			})
		}
	default:
		failures = append(failures, p.CheckFailure{
			Property: "teamType",
			Reason:   "must be either 'github' or 'pulumi'",
		})
	}
	return inputs, failures, nil
}

func (*Team) Delete(ctx p.Context, id string, props TeamState) error {
	return GetConfig(ctx).Client.DeleteTeam(ctx, props.OrganizationName, props.Name)
}

func (*Team) Read(
	ctx p.Context, id string, _ TeamInput, _ TeamState,
) (string, TeamInput, TeamState, error) {
	orgName, teamName, err := parseTeamId(id)
	if err != nil {
		return "", TeamInput{}, TeamState{}, err
	}

	team, err := GetConfig(ctx).Client.GetTeam(ctx, orgName, teamName)
	if err != nil || team == nil { // team == nil => the team was deleted
		return "", TeamInput{}, TeamState{}, err
	}

	inputs := TeamInput{
		Description:      team.Description,
		DisplayName:      team.DisplayName,
		Name:             team.Name,
		TeamType:         team.Type,
		OrganizationName: orgName,
	}
	for _, m := range team.Members {
		inputs.Members = append(inputs.Members, m.GithubLogin)
	}
	// Sort the members so the order is deterministic
	slices.Sort(inputs.Members)

	// TODO: We could be smart about allowing inputs.Members to be nil if the input
	// state has it as nil. This would allow better imports
	return id, inputs, TeamState{
		TeamInput: inputs,
		Members:   inputs.Members,
	}, nil

}

func (*Team) Update(ctx p.Context, id string, olds TeamState, news TeamInput, preview bool) (TeamState, error) {
	if preview {
		return TeamState{TeamInput: news}, nil
	}

	client := GetConfig(ctx).Client

	inputsChanged := olds
	if olds.Description != news.Description || olds.DisplayName != news.Description {

		inputsChanged.Description = news.Description
		inputsChanged.DisplayName = news.DisplayName

		err := client.UpdateTeam(ctx,
			inputsChanged.OrganizationName, inputsChanged.Name,
			inputsChanged.DisplayName, inputsChanged.Description)
		if err != nil {
			return TeamState{}, err
		}
	}

	// github teams can't manage membership.
	//
	// TODO: This should indicate a replace instead of simply ignoring the required
	// change.
	//
	// That can be done with a custom diff implementation.
	if !slices.Equal(olds.Members, news.Members) && news.TeamType != "github" {
		inputsChanged.Members = news.Members
		for _, usernameToDelete := range olds.Members {
			if !slices.Contains(news.Members, usernameToDelete) {
				// TODO: At this point, we should be returning partial
				// state unless DeleteMemberFromTeam is idempotent.
				err := client.DeleteMemberFromTeam(ctx,
					news.OrganizationName, news.Name, usernameToDelete)
				if err != nil {
					return TeamState{}, err
				}
			}
		}

		for _, usernameToAdd := range news.Members {
			if !slices.Contains(olds.Members, usernameToAdd) {
				// TODO: Likewise, we should ensure that we handle a
				// partial update unless AddMemberToTeam to idempotent.
				err := client.AddMemberToTeam(ctx, news.OrganizationName, news.Name, usernameToAdd)
				if err != nil {
					return TeamState{}, err
				}
			}
		}
	}

	return inputsChanged, nil
}

func (*Team) Create(
	ctx p.Context, name string, input TeamInput, preview bool,
) (string, TeamState, error) {
	if preview {
		return "", TeamState{TeamInput: input}, nil
	}
	client := GetConfig(ctx).Client
	team, err := client.CreateTeam(ctx,
		input.OrganizationName, input.Name,
		input.TeamType, input.DisplayName,
		input.Description, input.GithubTeamId)
	if err != nil {
		return "", TeamState{}, fmt.Errorf("error creating team '%s': %w", input.Name, err)
	}

	id := fmt.Sprintf("%s/%s", input.OrganizationName, team.Name)
	state := TeamState{TeamInput: input}

	// We have now created a team.
	//
	// It is very important to ensure that from this point on, any other error below
	// returns an [infer.ResourceInitFailedError] Otherwise, we leak a team resource.
	retErr := func(err error) (string, TeamState, error) {
		return id, state, infer.ResourceInitFailedError{
			Reasons: []string{err.Error()},
		}
	}

	for _, memberToAdd := range input.Members {
		err := client.AddMemberToTeam(ctx, input.OrganizationName, input.Name, memberToAdd)
		if err != nil {
			return retErr(err)
		}
		// if we've successfully added member to team, save them to the state we're going to return
		// so that a re-run can detect the left over members to add via Update
		state.Members = append(state.Members, memberToAdd)
	}

	return id, state, nil
}

// format: organization/teamName
func parseTeamId(id string) (string, string, error) {
	s := strings.Split(id, "/")
	if len(s) != 2 {
		return "", "", fmt.Errorf("%q is invalid, must contain a single slash ('/')", id)
	}
	return s[0], s[1], nil
}
