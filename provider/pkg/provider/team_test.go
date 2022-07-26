package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
)

func TestTeamResourceCreate(t *testing.T) {

	t.Run("Deletes Team After Failing To Add Members", func(t *testing.T) {
		// define input for team
		tInput := PulumiServiceTeamInput{
			Type:             "pulumi",
			Name:             "a-not-yet-created-team",
			OrganizationName: "org",
			Members: []string{
				"user-not-found",
			},
		}
		callNum := 0
		teamsPath := fmt.Sprintf("/api/orgs/%s/teams/%s", tInput.OrganizationName, tInput.Name)
		// define the calls we want to receive
		// we expect:
		//     POST to create team
		//     PATCH to add member <-- return 404 here for "user not found"
		//     DELETE to delete team <-- "roll back" the team creation due to 404
		wantCalls := []struct {
			ReqPath    string
			ReqMethod  string
			RespStatus int
			RespBody   interface{}
		}{
			{
				RespStatus: http.StatusCreated,
				// the team type is "pulumi", so that's why we POST to /teams/pulumi
				ReqPath:   fmt.Sprintf("/api/orgs/%s/teams/pulumi", tInput.OrganizationName),
				ReqMethod: http.MethodPost,
				RespBody:  pulumiapi.Team{},
			},
			{
				RespStatus: http.StatusNotFound,
				ReqPath:    teamsPath,
				ReqMethod:  http.MethodPatch,
			},
			{
				RespStatus: http.StatusOK,
				ReqPath:    teamsPath,
				ReqMethod:  http.MethodDelete,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// look up call and increment counter so we can look up the next call
			wantCall := wantCalls[callNum]
			callNum++

			// assert that the actual request matches what we want
			assert.Equal(t, wantCall.ReqMethod, r.Method)
			assert.Equal(t, wantCall.ReqPath, r.URL.Path, r.Method)
			w.WriteHeader(wantCall.RespStatus)
			// optionally write response body
			if wantCall.RespBody != nil {
				b, _ := json.Marshal(wantCall.RespBody)
				_, _ = w.Write(b)
			}
		}))
		defer server.Close()

		apiClient, err := pulumiapi.NewClient(server.Client(), "", server.URL)
		if err != nil {
			t.Fatal(err)
		}

		tr := &PulumiServiceTeamResource{
			client: apiClient,
		}

		props, err := plugin.MarshalProperties(tInput.ToPropertyMap(), plugin.MarshalOptions{})
		if err != nil {
			t.Fatal(err)
		}

		_, err = tr.Create(&pulumirpc.CreateRequest{
			Properties: props,
		})
		// because we 404 on the second call, Create should return an error
		assert.Error(t, err)

		// expect that we got all of the calls that we defined
		// validation of each call happens inside of server's HandlerFunc
		wantNumCalls := len(wantCalls)
		gotNumCalls := callNum
		assert.Equal(t, wantNumCalls, gotNumCalls)
	})
}
