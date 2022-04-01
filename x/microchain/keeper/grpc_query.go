package keeper

import (
	"github.com/mconcat/microchain/x/microchain/types"
)

var _ types.QueryServer = Keeper{}
