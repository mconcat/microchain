package keeper

import (
	"github.com/mconcat/microchain/x/permission/types"
)

var _ types.QueryServer = Keeper{}
