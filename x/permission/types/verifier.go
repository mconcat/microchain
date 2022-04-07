package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/capability/types"
	proto "github.com/gogo/protobuf/proto"
	ibctypes "github.com/cosmos/ibc-go/modules/core/exported"
)

// User send signature to the verifier, verifier retrieves appropriate State, verify using it.

// Corresponds to ibc.ClientState
// Signature corresponds to ibc.Packet+ibc.CommitmentProof pair
// Updates like CheckHeaderAndUpdateState is not handled here - should be handled by the concrete impls
// Verifier verifies whether a transaction is valid or not.
// Verifier can rely on the chain state, and update it.
// Verifying transaction itself should be able to be done without manipulating chain state(think checkTx)
// Verifier can hold assets to pay fee(so IBC relayers dont neet to pay fee)
// Once verifier validates the transaction, it returns a capability key that represent the valid owner.
// The capkey owner can be different from the verifier itself.
// Signature contains the authorization proof that can be verified by the verifier.
// Private key signature, IBC commitment proof, etc...
type Verifier[Sig Signature] interface {
	proto.Message

	GetAddress() sdk.AccAddress // Address of the verifier

	// Verify corresponds to
	// auth.AnteHandler
	// tendermint.VerifyPacketCommitment / tendermint.VerifyPacketAcknowledgement
	// localhost.VerifyPacketCommitment / localhost.VerifyPacketAcknowledgement
	Verify(ctx sdk.Context, sig Sig) (*types.Capability, error) // Verifier logic

	/*
		GetAsset() Asset // Assets that the verifier hold
		SetAsset(Asset)
	*/
}

type Signature interface {
	GetPortID() string
	GetChannelID() uint64
	GetHeight() uint64
	GetSequence() uint64
//	GetSignature() []byte // commitment proof
//	GetSignBytes() []byte // commitment bytes
}


type Actor[Pack Packet] interface {
	Address() sdk.AccAddress
	IsAllowedVerifier(sdk.AccAddress) bool

	Send(packet Pack) error
	Receive(packet Pack) error
	Ack(asset Asset, ack Acknowledgement[Pack]) error
}

// Correspond to ibc.Packet
// gRPC request
type Packet interface {
	proto.Message

	GetActorAddress() sdk.AccAddress
	GetAssets() 
}

// Correspond to ibc.Acknowledgement
// gRPC response
type Acknowledgement[Pack Packet] interface {
	proto.Message

	GetError() error
	GetPacket() Pack
}

// A gRPC service that represents an actor(module, contract, whatever)
// takes a packet and returns an acknowledgement.
// 