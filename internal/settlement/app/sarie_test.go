package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	settlementapp "github.com/mohsinazam/banking/internal/settlement/app"
)

func TestSarieMockAlwaysSucceedsWithZeroFailRate(t *testing.T) {
	mock := settlementapp.NewSarieMock(settlementapp.SarieConfig{
		FailRate: 0,
		MinDelay: time.Millisecond,
		MaxDelay: time.Millisecond,
	})

	result, err := mock.Settle(context.Background(), "stl_test_001")
	require.NoError(t, err)
	require.True(t, result.Success)
	require.Equal(t, "stl_test_001", result.Reference)
}

func TestSarieMockAlwaysFailsWithFullFailRate(t *testing.T) {
	mock := settlementapp.NewSarieMock(settlementapp.SarieConfig{
		FailRate: 1,
		MinDelay: time.Millisecond,
		MaxDelay: time.Millisecond,
	})

	result, err := mock.Settle(context.Background(), "stl_test_002")
	require.NoError(t, err)
	require.False(t, result.Success)
	require.NotEmpty(t, result.Error)
}
