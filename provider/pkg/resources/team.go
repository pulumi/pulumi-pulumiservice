package resources

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

type Team struct{}

var (
	_ infer.CustomCreate[TeamInput, TeamState] = &Team{}
	_ infer.CustomCheck[TeamInput]             = &Team{}
	_ infer.CustomDelete[TeamState]            = &Team{}
	_ infer.CustomRead[TeamInput, TeamState]   = &Team{}
	_ infer.CustomUpdate[TeamInput, TeamState] = &Team{}
)

func (t *Team) Annotate(a infer.Annotator) {
	a.Describe(t, "The Pulumi Cloud offers role-based access control (RBAC) using teams. Teams allow organization admins to assign a set of stack permissions to a group of users.")
}

type TeamCore struct {
	OrganizationName string   `pulumi:"organizationName" provider:"replaceOnChanges"`
	Type             string   `pulumi:"teamType" provider:"replaceOnChanges"`
	Name             *string  `pulumi:"name,optional" provider:"replaceOnChanges"`
	DisplayName      *string  `pulumi:"displayName,optional"`
	Description      *string  `pulumi:"description,optional"`
	GitHubTeamID     *float64 `pulumi:"githubTeamId,optional"`
}

func (t *TeamCore) Annotate(a infer.Annotator) {
	a.Describe(&t.Description, "Optional. Team description.")
	a.Describe(&t.DisplayName, "Optional. Team display name.")
	a.Describe(&t.Name, "The team's name. Required for \"pulumi\" teams.")
	a.Describe(&t.OrganizationName, "The name of the Pulumi organization the team belongs to.")
	a.Describe(&t.Type, "The type of team. Must be either `pulumi` or `github`.")
	a.Describe(&t.GitHubTeamID, `The GitHub ID of the team to mirror. Must be in the same GitHub organization that the Pulumi org is backed by. Required for "github" teams.`)
}

type TeamInput struct {
	TeamCore
	Members []string `pulumi:"members,optional"`
}

func (t *TeamInput) Annotate(a infer.Annotator) {
	a.Describe(&t.Members, "List of Pulumi Cloud usernames of team members.")
}

type TeamState struct {
	TeamCore
	Members []string `pulumi:"members"`
}

func (t *TeamState) Annotate(a infer.Annotator) {
	a.Describe(&t.Members, "List of Pulumi Cloud usernames of team members.")
}

func (*Team) Create(ctx context.Context, req infer.CreateRequest[TeamInput]) (infer.CreateResponse[TeamState], error) {
	teamURN := fmt.Sprintf("%s/%s", req.Inputs.OrganizationName, util.OrZero(req.Inputs.Name))
	if req.DryRun {
		return infer.CreateResponse[TeamState]{
			ID: teamURN,
			Output: TeamState{
				TeamCore: req.Inputs.TeamCore,
				Members:  req.Inputs.Members,
			},
		}, nil
	}
	client := config.GetClient(ctx)
	team, err := client.CreateTeam(ctx,
		req.Inputs.OrganizationName,
		util.OrZero(req.Inputs.Name),
		req.Inputs.Type,
		util.OrZero(req.Inputs.DisplayName),
		util.OrZero(req.Inputs.Description),
		int64(util.OrZero(req.Inputs.GitHubTeamID)),
	)
	if err != nil {
		return infer.CreateResponse[TeamState]{}, fmt.Errorf("error creating teamUrn '%s': %s", util.OrZero(req.Inputs.Name), err.Error())
	}

	// We have now created a teamUrn.  It is very important to ensure that from this point on, any other error
	// below returns the ID using the `pulumirpc.ErrorResourceInitFailed` error details annotation.  Otherwise,
	// we leak a teamUrn resource. We ensure that we wrap any errors in a partial error and return that to the RPC.

	members := []string{}
	for _, memberToAdd := range req.Inputs.Members {
		err = client.AddMemberToTeam(ctx, req.Inputs.OrganizationName, util.OrZero(req.Inputs.Name), memberToAdd)
		if err != nil {
			return infer.CreateResponse[TeamState]{
				ID: teamURN,
				Output: TeamState{
					TeamCore: req.Inputs.TeamCore,
					Members:  members,
				},
			}, infer.ResourceInitFailedError{Reasons: []string{err.Error()}}
		}
		members = append(members, memberToAdd)
	}

	// Outputs should be the result of a GetTeam call, so we can return the full teamUrn object with fidelity
	// including all new members that were added.
	team, err = client.GetTeam(ctx, req.Inputs.OrganizationName, team.Name)
	if err != nil {
		return infer.CreateResponse[TeamState]{
			ID: teamURN,
			Output: TeamState{
				TeamCore: req.Inputs.TeamCore,
				Members:  members,
			},
		}, infer.ResourceInitFailedError{Reasons: []string{err.Error()}}
	}
	members = members[:0]
	for _, v := range team.Members {
		members = append(members, v.GithubLogin)
	}
	// Sort the members so the order is deterministic
	slices.Sort(members)

	return infer.CreateResponse[TeamState]{
		ID: teamURN,
		Output: TeamState{
			Members: members,
			TeamCore: TeamCore{
				Description:      util.OrNil(team.Description),
				DisplayName:      util.OrNil(team.DisplayName),
				Name:             &team.Name,
				Type:             team.Type,
				OrganizationName: req.Inputs.OrganizationName,
				GitHubTeamID:     req.Inputs.GitHubTeamID,
			},
		},
	}, nil
}

func (*Team) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[TeamInput], error) {
	i, checkFailures, err := infer.DefaultCheck[TeamInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[TeamInput]{}, nil
	}
	if i.Type != "github" && i.Type != "pulumi" {
		checkFailures = append(checkFailures, p.CheckFailure{
			Reason:   fmt.Sprintf("found %q instead of 'pulumi' or 'github'", i.Type),
			Property: "type",
		})
	}

	if i.Type == "github" && i.GitHubTeamID == nil {
		checkFailures = append(checkFailures, p.CheckFailure{
			Reason:   "teams with teamType 'github' require a githubTeamId",
			Property: "githubTeamId",
		})
	}

	if i.Type == "pulumi" && i.Name == nil {
		checkFailures = append(checkFailures, p.CheckFailure{
			Reason:   "teams with teamType 'pulumi' require a name",
			Property: "name",
		})
	}

	if i.DisplayName == nil {
		i.DisplayName = i.Name
	}
	if i.Members == nil {
		i.Members = []string{}
	}
	slices.Sort(i.Members)

	return infer.CheckResponse[TeamInput]{
		Inputs:   i,
		Failures: checkFailures,
	}, nil
}

func (*Team) Delete(ctx context.Context, req infer.DeleteRequest[TeamState]) (infer.DeleteResponse, error) {
	client := config.GetClient(ctx)
	return infer.DeleteResponse{}, client.DeleteTeam(ctx, req.State.OrganizationName, util.OrZero(req.State.Name))
}

func (*Team) Read(ctx context.Context, req infer.ReadRequest[TeamInput, TeamState]) (infer.ReadResponse[TeamInput, TeamState], error) {
	client := config.GetClient(ctx)
	orgName, teamName, err := splitSingleSlashString(req.ID)
	if err != nil {
		return infer.ReadResponse[TeamInput, TeamState]{}, err
	}

	team, err := client.GetTeam(ctx, orgName, teamName)
	if err != nil {
		return infer.ReadResponse[TeamInput, TeamState]{}, fmt.Errorf("failed to read Team (%q): %w", req.ID, err)
	}
	if team == nil {
		return infer.ReadResponse[TeamInput, TeamState]{}, nil
	}

	core := TeamCore{
		OrganizationName: orgName,
		Type:             team.Type,
		Name:             &team.Name,
		DisplayName:      util.OrNil(team.DisplayName),
		Description:      util.OrNil(team.Description),
		GitHubTeamID:     req.Inputs.GitHubTeamID,
	}

	members := []string{}
	for _, m := range team.Members {
		members = append(members, m.GithubLogin)
	}
	slices.Sort(members)

	return infer.ReadResponse[TeamInput, TeamState]{
		ID: req.ID,
		Inputs: TeamInput{
			TeamCore: core,
			Members:  members,
		},
		State: TeamState{
			TeamCore: core,
			Members:  members,
		},
	}, nil
}

func (*Team) Update(ctx context.Context, req infer.UpdateRequest[TeamInput, TeamState]) (infer.UpdateResponse[TeamState], error) {
	if req.DryRun {
		return infer.UpdateResponse[TeamState]{
			Output: TeamState{
				TeamCore: req.Inputs.TeamCore,
				Members:  req.Inputs.Members,
			},
		}, nil
	}
	client := config.GetClient(ctx)

	if req.State.Description != req.Inputs.Description || req.State.DisplayName != req.Inputs.DisplayName {
		err := client.UpdateTeam(ctx, req.Inputs.OrganizationName, util.OrZero(req.Inputs.Name), util.OrZero(req.Inputs.DisplayName), util.OrZero(req.Inputs.Description))
		if err != nil {
			return infer.UpdateResponse[TeamState]{}, err
		}
	}

	// github teams can't manage membership.
	members := make([]string, len(req.State.Members))
	copy(members, req.State.Members)

	if !slices.Equal(req.Inputs.Members, req.State.Members) && req.Inputs.Type != "github" {
		for i := len(req.State.Members) - 1; i >= 0; i-- {
			usernameToDelete := req.State.Members[i]
			if !slices.Contains(req.Inputs.Members, usernameToDelete) {
				err := client.DeleteMemberFromTeam(ctx, req.Inputs.OrganizationName, util.OrZero(req.Inputs.Name), usernameToDelete)
				if err != nil {
					slices.Sort(members)
					// We have failed to delete a member, but we may
					// have still done something. Report on what we
					// did.
					return infer.UpdateResponse[TeamState]{
						Output: TeamState{
							TeamCore: req.Inputs.TeamCore,
							Members:  members,
						},
					}, infer.ResourceInitFailedError{Reasons: []string{err.Error()}}
				}
				// Remove this user from our running list of members
				if len(members) == 1 {
					// If there is only one member left, we zero out our list
					members = members[:0]
				} else {
					// Swap the current element with our last element, then pop the last element.
					members[i] = members[len(members)-1]
					members = members[:len(members)-1]
				}
			}
		}

		for _, usernameToAdd := range req.Inputs.Members {
			if !slices.Contains(req.State.Members, usernameToAdd) {
				err := client.AddMemberToTeam(ctx, req.Inputs.OrganizationName, util.OrZero(req.Inputs.Name), usernameToAdd)
				if err != nil {
					slices.Sort(members)
					return infer.UpdateResponse[TeamState]{
						Output: TeamState{
							TeamCore: req.Inputs.TeamCore,
							Members:  members,
						},
					}, infer.ResourceInitFailedError{Reasons: []string{err.Error()}}
				}
				members = append(members, usernameToAdd)
			}
		}
		slices.Sort(members)
		req.Inputs.Members = members
	}

	return infer.UpdateResponse[TeamState]{
		Output: TeamState{
			TeamCore: req.Inputs.TeamCore,
			Members:  req.Inputs.Members,
		},
	}, nil
}

func splitSingleSlashString(id string) (string, string, error) {
	// format: organization/webhookName
	s := strings.Split(id, "/")
	if len(s) != 2 {
		return "", "", fmt.Errorf("%q is invalid, must contain a single slash ('/')", id)
	}
	return s[0], s[1], nil
}
