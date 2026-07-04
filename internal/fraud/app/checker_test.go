package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	fraudredis "github.com/mohsinazam/banking/internal/fraud/adapters/redis"
	fraudapp "github.com/mohsinazam/banking/internal/fraud/app"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	"github.com/mohsinazam/banking/internal/transfer/ports"
)

func TestFraudRejectsOversizedTransfer(t *testing.T) {
	mr := miniredis.RunT(t)
	checker := fraudapp.NewChecker(
		fraudredis.NewVelocityStore(goredis.NewClient(&goredis.Options{Addr: mr.Addr()})),
		fraudapp.Config{MaxSingleHalalas: 100_00},
	)

	err := checker.Check(context.Background(), ports.FraudCheckInput{
		FromAccountID: "wallet-alice",
		Amount:        shareddomain.MustSAR(200, 0),
	})
	require.Error(t, err)
	require.True(t, shareddomain.IsDomainCode(err, shareddomain.ErrCodeForbidden))
}

func TestFraudRejectsVelocityLimit(t *testing.T) {
	mr := miniredis.RunT(t)
	store := fraudredis.NewVelocityStore(goredis.NewClient(&goredis.Options{Addr: mr.Addr()}))
	checker := fraudapp.NewChecker(store, fraudapp.Config{
		MaxHourlyTransfers: 2,
		VelocityWindow:     time.Hour,
	})

	ctx := context.Background()
	input := ports.FraudCheckInput{FromAccountID: "wallet-alice", Amount: shareddomain.MustSAR(10, 0)}
	require.NoError(t, checker.Check(ctx, input))
	require.NoError(t, checker.Check(ctx, input))

	err := checker.Check(ctx, input)
	require.Error(t, err)
	require.True(t, shareddomain.IsDomainCode(err, shareddomain.ErrCodeForbidden))
}
