package base

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/mconcat/microchain/x/permission/types"
)

// BaseAccount defines privkey based account, holding tokens
var _ types.Verifier[BaseAccountSignature] = &BaseAccount{}

type BaseAccount struct {
	authtypes.BaseAccount
}

// GetSignerAcc returns an account for a given address that is expected to sign
// a transaction.
func GetSignerAcc(ctx sdk.Context, addr sdk.AccAddress) (authtypes.AccountI, error) {
	if acc := ak.GetAccount(ctx, addr); acc != nil {
		return acc, nil
	}

	return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
}

// OnlyLegacyAminoSigners checks SignatureData to see if all
// signers are using SIGN_MODE_LEGACY_AMINO_JSON. If this is the case
// then the corresponding SignatureV2 struct will not have account sequence
// explicitly set, and we should skip the explicit verification of sig.Sequence
// in the SigVerificationDecorator's AnteHandler function.
func OnlyLegacyAminoSigners(sigData signing.SignatureData) bool {
	switch v := sigData.(type) {
	case *signing.SingleSignatureData:
		return v.SignMode == signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	case *signing.MultiSignatureData:
		for _, s := range v.Signatures {
			if !OnlyLegacyAminoSigners(s) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

type BaseAccountSignature struct {
	signing.SignatureV2
	SignBytes []byte
}

func (sig BaseAccountSignature) GetPortID() string { return "account" }
func (sig BaseAccountSignature) GetChannelID() uint64 { return 0 }
func (sig BaseAccountSignature) GetSequence() uint64 { return sig.Sequence }
func (sig BaseAccountSignature) GetHeight() uint64 { return 0 }
// func (sig BaseAccountSignature) GetSignature() []byte { return sig. }
// func (sig BaseAccountSignature) GetSignBytes() []byte { return sig.SignBytes }


func (acc BaseAccount) MakeSignature(ctx sdk.Context, handler authsigning.SignModeHandler, tx sdk.Tx) (BaseAccountSignature, error) {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return BaseAccountSignature{}, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return BaseAccountSignature{}, err
	}

	// signerAddrs should be already retrieved somewhere else
	// signerAddrs := sigTx.GetSigners()

	// TODO: support multisig transaction by defining multiaccount
	if len(sigs) != 1 {
		return BaseAccountSignature{}, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "BaseAccount supports only one signature")
	}

	sig := sigs[0]

	// retrieve signer data
	genesis := ctx.BlockHeight() == 0
	var accNum uint64
	if !genesis {
		accNum = acc.GetAccountNumber()
	}

	signerData := authsigning.SignerData{
		ChainID:       ctx.ChainID(),
		AccountNumber: accNum,
		Sequence:      acc.GetSequence(),
	}

	signBytes, err := MakeSignModeHandler().GetSignBytes(sig.Data.(*signing.SingleSignatureData).SignMode, signerData, tx)
	if err != nil {
		return BaseAccountSignature{}, err
	}

	return BaseAccountSignature{
		SignatureV2: sig,
		SignBytes:   signBytes,
	}, nil
}

func (acc BaseAccount) Verify(ctx sdk.Context, sig BaseAccountSignature) error {
	// no need to verify signatures on recheck tx
	if ctx.IsReCheckTx() {
		return nil
	}

	// check that signer length and signature length are the same
	/*
		if len(sigs) != len(signerAddrs) {
			return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "invalid number of signer;  expected: %d, got %d", len(signerAddrs), len(sigs))
		}
	*/

	// retrieve pubkey
	pubKey := acc.GetPubKey()
	if /*!simulate &&*/ pubKey == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, "pubkey on account is not set")
	}


	// Check account sequence number.
	if sig.Sequence != acc.GetSequence() {
		return sdkerrors.Wrapf(
			sdkerrors.ErrWrongSequence,
			"account sequence mismatch, expected %d, got %d", acc.GetSequence(), sig.Sequence,
		)
	}

	//if !simulate {
	if !pubKey.VerifySignature(sig.SignBytes, sig.Data.(*signing.SingleSignatureData).Signature) {
		var errMsg string
		if OnlyLegacyAminoSigners(sig.Data) {
			// If all signers are using SIGN_MODE_LEGACY_AMINO, we rely on VerifySignature to check account sequence number,
			// and therefore communicate sequence number as a potential cause of error.
			errMsg = fmt.Sprintf("signature verification failed; please verify account number (%d), sequence (%d)", acc.AccountNumber, acc.Sequence)
		} else {
			errMsg = fmt.Sprintf("signature verification failed; please verify account number (%d)", acc.AccountNumber)
		}
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, errMsg)

	}
	//}
	//	}

	return nil
}
