package keeper

import (
	"github.com/mconcat/microchain/x/consensus/types"
)

var _ types.QueryServer = Keeper{}
