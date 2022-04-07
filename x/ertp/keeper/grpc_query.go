package keeper

import (
	"github.com/mconcat/microchain/x/ertp/types"
)

var _ types.QueryServer = Keeper{}
