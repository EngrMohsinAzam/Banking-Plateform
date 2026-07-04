package auth_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mohsinazam/banking/internal/platform/auth"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
)

func TestAPIKeyDisabledAllowsAll(t *testing.T) {
	v := auth.NewAPIKeyValidator(false, []string{"secret"})
	require.NoError(t, v.Validate(""))
}

func TestAPIKeyEnabledRejectsMissing(t *testing.T) {
	v := auth.NewAPIKeyValidator(true, []string{"secret"})
	err := v.Validate("")
	require.Error(t, err)
	require.True(t, shareddomain.IsDomainCode(err, shareddomain.ErrCodeForbidden))
}

func TestAPIKeyEnabledAcceptsValid(t *testing.T) {
	v := auth.NewAPIKeyValidator(true, []string{"secret", "other"})
	require.NoError(t, v.Validate("secret"))
}

func TestAPIKeyEnabledRejectsInvalid(t *testing.T) {
	v := auth.NewAPIKeyValidator(true, []string{"secret"})
	err := v.Validate("wrong")
	require.Error(t, err)
}
