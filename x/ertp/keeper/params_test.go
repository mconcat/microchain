package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "github.com/mconcat/microchain/testutil/keeper"
	"github.com/mconcat/microchain/x/ertp/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.ErtpKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
