package resources

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitOrgAccessTokenId(t *testing.T) {
	t.Run("Splits org access token id", func(t *testing.T) {
		tokenID := "org/name/id"

		org, name, id, err := splitOrgAccessTokenID(tokenID)
		assert.NoError(t, err)

		assert.Equal(t, "org", org)
		assert.Equal(t, "name", name)
		assert.Equal(t, "id", id)
	})

	t.Run("Splits org access token id with name with slashes", func(t *testing.T) {
		tokenID := "org/name/with/slashes/id" //nolint:gosec // This is test data, not a credential

		org, name, id, err := splitOrgAccessTokenID(tokenID)
		assert.NoError(t, err)

		assert.Equal(t, "org", org)
		assert.Equal(t, "name/with/slashes", name)
		assert.Equal(t, "id", id)
	})

	t.Run("Splits org access token id with invalid id", func(t *testing.T) {
		tokenID := "org/badname"

		_, _, _, err := splitOrgAccessTokenID(tokenID)
		assert.ErrorContains(t, err, fmt.Sprintf("%q is invalid, must contain a single slash ('/')", tokenID))
	})
}
