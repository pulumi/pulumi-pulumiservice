package resources

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitOrgAccessTokenId(t *testing.T) {
	t.Run("Splits org access token id", func(t *testing.T) {
		tokenId := "org/name/id"

		org, name, id, err := splitOrgAccessTokenId(tokenId)
		assert.NoError(t, err)

		assert.Equal(t, "org", org)
		assert.Equal(t, "name", name)
		assert.Equal(t, "id", id)
	})

	t.Run("Splits org access token id with name with slashes", func(t *testing.T) {
		tokenId := "org/name/with/slashes/id"

		org, name, id, err := splitOrgAccessTokenId(tokenId)
		assert.NoError(t, err)

		assert.Equal(t, "org", org)
		assert.Equal(t, "name/with/slashes", name)
		assert.Equal(t, "id", id)
	})

	t.Run("Splits org access token id with invalid id", func(t *testing.T) {
		tokenId := "org/badname"

		_, _, _, err := splitOrgAccessTokenId(tokenId)
		assert.ErrorContains(t, err, fmt.Sprintf("%q is invalid, must contain a single slash ('/')", tokenId))
	})
}
