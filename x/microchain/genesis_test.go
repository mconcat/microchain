package microchain_test

import (
	"testing"

	keepertest "github.com/mconcat/microchain/testutil/keeper"
	"github.com/mconcat/microchain/testutil/nullify"
	"github.com/mconcat/microchain/x/microchain"
	"github.com/mconcat/microchain/x/microchain/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.MicrochainKeeper(t)
	microchain.InitGenesis(ctx, *k, genesisState)
	got := microchain.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
