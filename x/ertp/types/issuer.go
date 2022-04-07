package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Asset represents arbitrary non-doublespendable data onchain.
type Purse interface {
	proto.Message

	PurseType() string
	// GetID() returns the globally unique immutable ID for this asset.
	// Capability owner for this ID owns this asset.
	GetID() uint64
	// Spend() splits the existing asset into multiple assets.
	// The new assets can have different type, representing state transition.
	Withdraw(instruction any) (Purse, Payment, error)
	// Transfer() transfers ownership of this asset.
	// It first requests transfer approval from `from`, and transfers to `to` by calling receive callback.
	// The receiving side could fail - in this case, the `from`'s fallback should be called.
	Deposit(pay Payment) (Purse, error) // change ownership
}

type TokenPurse struct {
	Amount uint64
}

func (purse TokenPurse) Withdraw(instruction any) (Purse, Payment, error) {
	switch instruction.(type) {
	case TokenSpendInstruction:
		return TokenPurse{purse.Amount-instruction.Amount}, 
	}
}

type Issuer interface {
	GetAddress() sdk.AccAddress // issuer.getAllegedName
	GetAmountOf(payment Payment) uint64 // issuer.getAmountOf
	GetDenom() string // issuer.getBrand
	MakeEmptyPurse() Purse // issuer.makeEmptyPurse
	Burn(payment Payment, expectedAmount uint64) uint64 // issuer.burn
	Claim(payment Payment, expectedAmount uint64) payment // issuer.claim
	Combine(payments []Payment, totalAmount uint64) Payment // issuer.combine
	Split(payment Payment, amount ...uint64) []Payment // issuer.split(Many)
}

type Mint interface {
	
}
