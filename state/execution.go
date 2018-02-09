package state

import (
	"time"
	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/events"

	"github.com/lino-network/lino/types"
	ttx "github.com/lino-network/lino/types/tx"
)

// If the tx is invalid, a TMSP error will be returned.
func ExecTx(state *State, pgz *types.Plugins, tx ttx.Tx, isCheckTx bool, evc events.Fireable) abci.Result {
	chainID := state.GetChainID()

	// Exec tx
	switch tx := tx.(type) {
	case *ttx.SendTx:
		// Validate inputs and outputs, basic
		res := tx.Input.ValidateBasic()
		if res.IsErr() {
			return res.PrependLog("in Input validate basic")
		}
		res = tx.Output.ValidateBasic()
		if res.IsErr() {
			return res.PrependLog("in Output validate basic")
		}

		// Get input account
		inAcc := state.GetAccount(tx.Input.Username)

		if inAcc == nil {
			// TODO change to UnknownUsername
			return abci.ErrBaseUnknownAddress
		}
		// Get output account
		outAcc := state.GetAccount(tx.Output.Username)

		if outAcc == nil {
			// TODO change to UnknownUsername
			return abci.ErrBaseUnknownAddress
		}

		// Validate inputs and outputs, advanced
		signBytes := tx.SignBytes(chainID)
		res = validateInputAdvanced(inAcc, signBytes, tx.Input)
		if res.IsErr() {
			return res.PrependLog("in validateInputAdvanced()")
		}
		fees := types.Coins{tx.Fee}
		outPlusFees := tx.Output.Coins
		if fees.IsValid() { // TODO: fix coins.Plus()
			outPlusFees = outPlusFees.Plus(fees)
		}
		if !tx.Input.Coins.IsEqual(outPlusFees) {
			return abci.ErrBaseInvalidOutput.AppendLog(cmn.Fmt("Input total (%v) != output total + fees (%v)", tx.Input.Coins, outPlusFees))
		}

		// TODO: Fee validation for SendTx

		// Good! Adjust accounts
		adjustByInput(state, inAcc, tx.Input)
		adjustByOutput(state, outAcc, tx.Output, isCheckTx)

		/*
			// Fire events
			if !isCheckTx {
				if evc != nil {
					for _, i := range tx.Inputs {
						evc.FireEvent(types.EventStringAccInput(i.Address), types.EventDataTx{tx, nil, ""})
					}
					for _, o := range tx.Outputs {
						evc.FireEvent(types.EventStringAccOutput(o.Address), types.EventDataTx{tx, nil, ""})
					}
				}
			}
		*/

		return abci.NewResultOK(ttx.TxID(chainID, tx), "")

	case *ttx.AppTx:
		// Validate input, basic
		res := tx.Input.ValidateBasic()
		if res.IsErr() {
			return res
		}

		// Get input account
		inAcc := state.GetAccount(tx.Input.Username)
		if inAcc == nil {
			return abci.ErrBaseUnknownAddress
		}

		// Validate input, advanced
		signBytes := tx.SignBytes(chainID)
		res = validateInputAdvanced(inAcc, signBytes, tx.Input)
		if res.IsErr() {
			state.logger.Info(cmn.Fmt("validateInputAdvanced failed on %X: %v", tx.Input.Username, res))
			return res.PrependLog("in validateInputAdvanced()")
		}
		if !tx.Input.Coins.IsGTE(types.Coins{tx.Fee}) {
			state.logger.Info(cmn.Fmt("Sender did not send enough to cover the fee %X", tx.Input.Username))
			return abci.ErrBaseInsufficientFunds.AppendLog(cmn.Fmt("input coins is %v, but fee is %v", tx.Input.Coins, types.Coins{tx.Fee}))
		}

		// Validate call address
		plugin := pgz.GetByName(tx.Name)
		if plugin == nil {
			return abci.ErrBaseUnknownAddress.AppendLog(
				cmn.Fmt("Unrecognized plugin name%v", tx.Name))
		}

		// Good!
		coins := tx.Input.Coins.Minus(types.Coins{tx.Fee})
		inAcc.LastTransaction += 1
		inAcc.Balance = inAcc.Balance.Minus(tx.Input.Coins)

		// If this is a CheckTx, stop now.
		if isCheckTx {
			state.SetAccount(tx.Input.Username, inAcc)
			return abci.OK
		}

		// Create inAcc checkpoint
		inAccCopy := inAcc.Copy()

		// Run the tx.
		cache := state.CacheWrap()
		cache.SetAccount(tx.Input.Username, inAcc)
		ctx := types.NewCallContext(tx.Input.Username, inAcc, coins)
		res = plugin.RunTx(cache, ctx, tx.Data)
		if res.IsOK() {
			cache.CacheSync()
			state.logger.Info("Successful execution")
			// Fire events
			/*
				if evc != nil {
					exception := ""
					if res.IsErr() {
						exception = res.Error()
					}
					evc.FireEvent(types.EventStringAccInput(tx.Input.Address), types.EventDataTx{tx, ret, exception})
					evc.FireEvent(types.EventStringAccOutput(tx.Address), types.EventDataTx{tx, ret, exception})
				}
			*/
		} else {
			state.logger.Info("AppTx failed", "error", res)
			// Just return the coins and return.
			inAccCopy.Balance = inAccCopy.Balance.Plus(coins)
			// But take the gas
			// TODO
			state.SetAccount(tx.Input.Username, inAccCopy)
		}
		return res

	case *ttx.PostTx:
		res := tx.ValidateBasic()
		if res.IsErr() {
			return res
		}
		// Get post author account
		acc := state.GetAccount(tx.Author)
		if acc == nil {
			return abci.ErrBaseUnknownAddress.AppendLog(
				cmn.Fmt("Unrecognized username%v", tx.Author))
		}

		signBytes := tx.SignBytes(chainID)
		res = validatePostAdvanced(acc, signBytes, *tx)
		if res.IsErr() {
			state.logger.Info(cmn.Fmt("validatePostAdvanced failed on %X: %v", tx.Author, res))
			return res.PrependLog("in validatePostAdvanced()")
		}
		post := &types.Post {
			Title: tx.Title,
			Content: tx.Content,
			Author: tx.Author,
			Sequence: tx.Sequence,
			Parent: tx.Parent,
			Created: time.Now(),
			LastUpdate: time.Now(),
			LastActivity: time.Now(),
			AllowReplies: true,
			AllowVotes: true,
		}

		var parentPost *types.Post
		if len(tx.Parent) > 0 {
			parentPost = state.GetPost(tx.Parent)
			if parentPost == nil {
				// TODO change to unknownpost
				return abci.ErrBaseUnknownAddress
			}
		}
		state.PostTxUpdateState(post, acc, parentPost)
		return abci.NewResultOK(ttx.TxID(chainID, tx), "")

	case *ttx.DonateTx:
		res := tx.ValidateBasic()
		if res.IsErr() {
			return res
		}
        acc := state.GetAccount(tx.Input.Username)
		if acc == nil {
			return abci.ErrBaseUnknownAddress
		}

		// Validate input, advanced
		signBytes := tx.SignBytes(chainID)
		res = validateInputAdvanced(acc, signBytes, tx.Input)
		if res.IsErr() {
			state.logger.Info(cmn.Fmt("validateInputAdvanced failed on %X: %v", tx.Input.Username, res))
			return res.PrependLog("in validateInputAdvanced()")
		}
		if !tx.Input.Coins.IsGTE(types.Coins{tx.Fee}) {
			state.logger.Info(cmn.Fmt("Sender did not send enough to cover the fee %X", tx.Input.Username))
			return abci.ErrBaseInsufficientFunds.AppendLog(cmn.Fmt("input coins is %v, but fee is %v", tx.Input.Coins, tx.Fee))
		}

		post := state.GetPost(tx.To)
		if post == nil {
			// TODO change to unknown post error
			return abci.ErrBaseUnknownAddress
		}

		acc.Balance = acc.Balance.Minus(tx.Input.Coins)
		acc.LastTransaction += 1
		state.SetAccount(tx.Input.Username, acc)
		state.DonateTxUpdateState(post, acc, tx.Input.Coins)
		if isCheckTx {
			return abci.OK
		}

		outAcc := state.GetAccount(post.Author)
		if outAcc == nil {
			// TODO change to unknown post
			return abci.ErrBaseUnknownAddress
		}
		outCoin := tx.Input.Coins.Minus(types.Coins{tx.Fee})
		outAcc.Balance = outAcc.Balance.Plus(outCoin)
		state.SetAccount(post.Author, outAcc)
		return abci.NewResultOK(ttx.TxID(chainID, tx), "")

	case *ttx.LikeTx:
		res := tx.ValidateBasic()
		if res.IsErr() {
			return res
		}
		account := state.GetAccount(tx.From)
		if account == nil {
			return abci.ErrBaseUnknownAddress
		}
		signBytes := tx.SignBytes(chainID)
		res = validateLikeAdvanced(account, signBytes, *tx)
		// Get post author account
		if res.IsErr() {
			state.logger.Info(cmn.Fmt("validateLikeAdvanced failed on %X: %v", tx.From, res))
			return res.PrependLog("in validateLikeAdvanced()")
		}
		post := state.GetPost(tx.To)
		if post == nil {
			// TODO change to unknown Post
			return abci.ErrBaseUnknownAddress
		}
		like := &types.Like{
			From   : tx.From,
			To     : tx.To,
			Weight : tx.Weight,
		}
		state.LikeTxUpdateState(like, account, post)
		return abci.NewResultOK(ttx.TxID(chainID, tx), "")

	case *ttx.FollowTx:
		res := tx.ValidateBasic()
		if res.IsErr() {
			return res
		}
		followerAcc := state.GetAccount(tx.Follower)
		if followerAcc == nil {
			return abci.ErrBaseUnknownAddress
		}
		signBytes := tx.SignBytes(chainID)
		res = validateFollowAdvanced(followerAcc, signBytes, *tx)
		// Get post author account
		
		followingAcc := state.GetAccount(tx.Following)
		if followingAcc == nil {
			return abci.ErrBaseUnknownAddress
		}
		if (tx.IsFollow) {
			state.FollowTxUpdateState(followerAcc, followingAcc)
		} else {
			state.UnfollowTxUpdateState(followerAcc, followingAcc)
		}
		return abci.NewResultOK(ttx.TxID(chainID, tx), "")
	default:
		return abci.ErrBaseEncodingError.SetLog("Unknown tx type")
	}
}

func validateInputAdvanced(acc *types.Account, signBytes []byte, in ttx.TxInput) (res abci.Result) {
	// Check sequence/coins
	seq, balance := acc.LastTransaction, acc.Balance
	if seq+1 != in.Sequence {
		return abci.ErrBaseInvalidSequence.AppendLog(cmn.Fmt("Got %v, expected %v. (acc.seq=%v)", in.Sequence, seq+1, acc.LastTransaction))
	}
	// Check amount
	if !balance.IsGTE(in.Coins) {
		return abci.ErrBaseInsufficientFunds.AppendLog(cmn.Fmt("balance is %v, tried to send %v", balance, in.Coins))
	}
	// Check signatures
	if !acc.PubKey.VerifyBytes(signBytes, in.Signature) {
		return abci.ErrBaseInvalidSignature.AppendLog(cmn.Fmt("SignBytes: %X", signBytes))
	}
	return abci.OK
}

func validatePostAdvanced(acc *types.Account, signBytes []byte, post ttx.PostTx) (res abci.Result) {
	// Check sequence
	seq := acc.LastPost
	if seq + 1 != post.Sequence {
		return abci.ErrBaseInvalidSequence.AppendLog(cmn.Fmt("Got %v, expected %v. (acc.seq=%v)", post.Sequence, seq+1, acc.LastPost))
	}
	// Check signatures
	if !acc.PubKey.VerifyBytes(signBytes, post.Signature) {
		return abci.ErrBaseInvalidSignature.AppendLog(cmn.Fmt("SignBytes: %X", signBytes))
	}
	return abci.OK
}

func validateLikeAdvanced(acc *types.Account, signBytes []byte, like ttx.LikeTx) (res abci.Result) {
	// Check signatures
	if !acc.PubKey.VerifyBytes(signBytes, like.Signature) {
		return abci.ErrBaseInvalidSignature.AppendLog(cmn.Fmt("SignBytes: %X", signBytes))
	}
	return abci.OK
}

func validateFollowAdvanced(acc *types.Account, signBytes []byte, ftx ttx.FollowTx) (res abci.Result) {
	// Check signatures
	if !acc.PubKey.VerifyBytes(signBytes, ftx.Signature) {
		return abci.ErrBaseInvalidSignature.AppendLog(cmn.Fmt("SignBytes: %X", signBytes))
	}
	return abci.OK
}

func adjustByInput(state types.AccountSetter, inAcc *types.Account, in ttx.TxInput) {
	if !inAcc.Balance.IsGTE(in.Coins) {
		cmn.PanicSanity("adjustByInputs() expects sufficient funds")
	}
	inAcc.Balance = inAcc.Balance.Minus(in.Coins)
	inAcc.LastTransaction += 1
	state.SetAccount(in.Username, inAcc)
}

func adjustByOutput(state *State, outAcc *types.Account, out ttx.TxOutput, isCheckTx bool) {
	outAcc.Balance = outAcc.Balance.Plus(out.Coins)
	if !isCheckTx {
		state.SetAccount(out.Username, outAcc)
	}
}
