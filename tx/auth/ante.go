package auth

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lino-network/lino/types"
)

// NewAnteHandler creates a new ante handler used for all tx
func NewAnteHandler(am types.AccountManager) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx,
	) (_ sdk.Context, _ sdk.Result, abort bool) {
		// Assert that there are signatures.
		var sigs = tx.GetSignatures()
		if len(sigs) == 0 {
			return ctx,
				sdk.ErrUnauthorized("no signers").Result(),
				true
		}

		// TODO: can tx just implement message?
		msg := tx.GetMsg()

		// Assert that number of signatures is correct.
		var signerAddrs = msg.GetSigners()
		if len(sigs) != len(signerAddrs) {
			return ctx,
				sdk.ErrUnauthorized("wrong number of signers").Result(),
				true
		}

		// Collect accounts to set in the context
		var signerAccs = make([]types.Account, len(signerAddrs))

		// Get the sign bytes by collecting all sequence numbers
		sequences := make([]int64, len(signerAddrs))
		for i := 0; i < len(signerAddrs); i++ {
			sequences[i] = sigs[i].Sequence
		}
		signBytes := sdk.StdSignBytes(ctx.ChainID(), sequences, msg)

		// Check sig and sequence for all signers
		for i := 0; i < len(sigs); i++ {
			signerAddr, sig := signerAddrs[i], sigs[i]

			signerAccKey := types.AccountKey(signerAddr)
			signerAcc, res := processSig(ctx, am, signerAccKey, sig, signBytes)
			if !res.IsOK() {
				return ctx, res, true
			}
			signerAccs[i] = signerAcc
		}
		ctx = WithSigners(ctx, signerAccs)
		return ctx, sdk.Result{}, false // continue...
	}
}

func processSig(ctx sdk.Context, am types.AccountManager, accKey types.AccountKey, sig sdk.StdSignature, signBytes []byte) (acc types.Account, res sdk.Result) {
	// Get the account
	acc = am.GetAccount(ctx, accKey)
	if acc == nil {
		return nil, ErrUnrecognizedUsername(accKey).Result()
	}

	// Verify and increment sequence number
	seq := acc.GetSequence()
	if seq != sig.Sequence {
		return nil, sdk.ErrInvalidSequence(
			fmt.Sprintf("Invalid sequence. Got %d, expected %d", sig.Sequence, seq)).Result()
	}
	acc.SetSequence(seq + 1)

	// Check and possibly set pubkey
	pubKey := acc.GetOwnerKey()
	if pubKey.Empty() {
		// pubKey could be nil
		pubKey = sig.PubKey
		err := acc.SetOwnerKey(pubKey)
		if err != nil {
			// TODO: Incorporate lower level err?
			return nil, sdk.ErrInternal("setting PubKey on signer").Result()
		}
	}

	// TODO: maybe enfore pubKey == sig.PubKey?

	// Check sig
	if !sig.PubKey.VerifyBytes(signBytes, sig.Signature) {
		return nil, sdk.ErrUnauthorized("signature verification failed").Result()
	}

	// Save the account
	am.SetAccount(ctx, acc)
	// TODO: result is not set?
	return
}
