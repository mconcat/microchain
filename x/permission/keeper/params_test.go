package keeper_test

import (
	"testing"

	testkeeper "github.com/mconcat/microchain/testutil/keeper"
	"github.com/mconcat/microchain/x/permission/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.PermissionKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
