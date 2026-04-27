package pulumiapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrentUser(t *testing.T) {
	c := startTestServer(t, testServerConfig{
		ExpectedReqMethod: http.MethodGet,
		ExpectedReqPath:   "/api/user",
		ResponseCode:      200,
		ResponseBody: CurrentUser{
			ID:          "user-123",
			GithubLogin: "alice",
			Name:        "Alice Example",
			Email:       "alice@example.com",
			AvatarURL:   "https://avatars.example.com/alice.png",
		},
	})
	got, err := c.GetCurrentUser(ctx)
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Equal(t, "alice", got.GithubLogin)
		assert.Equal(t, "Alice Example", got.Name)
		assert.Equal(t, "alice@example.com", got.Email)
	}
}
