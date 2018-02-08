package state

import (
	"github.com/lino-network/lino/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tmlibs/log"
	"testing"
	abci "github.com/tendermint/abci/types"
	ttx "github.com/lino-network/lino/types/tx"
)

//--------------------------------------------------------
// test environment is a bunch of lists of accountns

type execTest struct {
	chainID string
	store   types.KVStore
	state   *State
	accIn   types.PrivAccount
	accOut  types.PrivAccount
}

func newExecTest() *execTest {
	et := &execTest{
		chainID: "test_chain_id",
	}
	et.reset()
	return et
}

func (et *execTest) signTx(tx *ttx.SendTx, accIn types.PrivAccount) {
	signBytes := tx.SignBytes(et.chainID)
	tx.Input.Signature = accIn.Sign(signBytes)
}

// returns the final balance and expected balance for input and output accounts
func (et *execTest) exec(tx *ttx.SendTx, checkTx bool) (res abci.Result, inGot, inExp, outGot, outExp types.Coins) {
	initBalIn := et.state.GetAccount(et.accIn.Account.Username).Balance
	initBalOut := et.state.GetAccount(et.accOut.Account.Username).Balance

	res = ExecTx(et.state, nil, tx, checkTx, nil)

	endBalIn := et.state.GetAccount(et.accIn.Account.Username).Balance
	endBalOut := et.state.GetAccount(et.accOut.Account.Username).Balance
	decrBalInExp := tx.Output.Coins.Plus(types.Coins{tx.Fee}) //expected decrease in balance In
	return res, endBalIn, initBalIn.Minus(decrBalInExp), endBalOut, initBalOut.Plus(tx.Output.Coins)
}

func (et *execTest) acc2State(privAccs ...types.PrivAccount) {
	for _, privAcc := range privAccs {
		et.state.SetAccount(privAcc.Account.Username, &privAcc.Account)
	}
}

//reset everything. state is empty
func (et *execTest) reset() {
	et.accIn = ttx.MakeAcc("foo", "foo")
	et.accOut = ttx.MakeAcc("bar", "bar")

	et.store = types.NewMemKVStore()
	et.state = NewState(et.store)
	et.state.SetLogger(log.TestingLogger())
	et.state.SetChainID(et.chainID)

	// NOTE we dont run acc2State here
	// so we can test non-existing accounts
}

//--------------------------------------------------------

func TestValidateInputAdvanced(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//validate input advanced
	tx := ttx.MakeSendTx(1, et.accOut, et.accIn)

	et.acc2State(et.accIn, et.accOut)
	signBytes := tx.SignBytes(et.chainID)

	//unsigned case
	res := validateInputAdvanced(&et.accIn.Account, signBytes, tx.Input)
	assert.True(res.IsErr(), "validateInputAdvanced: expected error on tx input without signature")

	//good signed case
	et.signTx(tx, et.accIn)
	res = validateInputAdvanced(&et.accIn.Account, signBytes, tx.Input)
	assert.True(res.IsOK(), "validateInputAdvanced: expected no error on good tx input. Error: %v", res.Error())

	//bad sequence case
	et.accIn.LastTransaction = 1
	et.signTx(tx, et.accIn)
	res = validateInputAdvanced(&et.accIn.Account, signBytes, tx.Input)
	assert.Equal(abci.CodeType_BaseInvalidSequence, res.Code, "validateInputAdvanced: expected error on tx input with bad sequence")
	et.accIn.LastTransaction = 0 //restore sequence

	//bad balance case
	et.accIn.Balance = types.Coins{{"mycoin", 2}}
	et.signTx(tx, et.accIn)
	res = validateInputAdvanced(&et.accIn.Account, signBytes, tx.Input)
	assert.Equal(abci.CodeType_BaseInsufficientFunds, res.Code,
		"validateInputAdvanced: expected error on tx input with insufficient funds %v", et.accIn.LastTransaction)
}

func TestValidatePostAdvanced(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	acc := ttx.MakeAcc("post", "post")

	//validatePostAdvanced
	tx := ttx.MakePostTx(acc.Account.Username, 1, acc)
	signBytes := tx.SignBytes(et.chainID)

	//unsigned case
	res := validatePostAdvanced(&acc.Account, signBytes, *tx)
	assert.True(res.IsErr(), "validateInputAdvanced: expected error on tx input without signature")

	//good signed case
	tx.Signature = acc.Sign(signBytes)
	res = validatePostAdvanced(&acc.Account, signBytes, *tx)
	assert.True(res.IsOK(), "validateInputAdvanced: expected no error on good tx input. Error: %v", res.Error())

	//bad sequence case
	acc.LastPost = 1
	tx.Signature = acc.Sign(signBytes)
	res = validatePostAdvanced(&acc.Account, signBytes, *tx)
	assert.Equal(abci.CodeType_BaseInvalidSequence, res.Code, "validateInputAdvanced: expected error on tx input with bad sequence")
}

func TestAdjustBy(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//adjustByInputs/adjustByOutputs
	//sending transaction from accIn to accOut
	initBalIn := et.accIn.Account.Balance
	initBalOut := et.accOut.Account.Balance
	et.acc2State(et.accIn, et.accOut)

	txIn := ttx.Acc2TxInput(1, et.accIn)
	txOut := ttx.Acc2TxOutput(et.accOut)

	adjustByInput(et.state, &et.accIn.Account, txIn)
	adjustByOutput(et.state, &et.accOut.Account, txOut, false)

	endBalIn := et.accIn.Account.Balance
	endBalOut := et.accOut.Account.Balance
	decrBalIn := initBalIn.Minus(endBalIn)
	incrBalOut := endBalOut.Minus(initBalOut)

	assert.True(decrBalIn.IsEqual(txIn.Coins),
		"adjustByInputs: total coins are not equal. diff: %v, tx: %v", decrBalIn.String(), txIn.Coins.String())
	assert.True(incrBalOut.IsEqual(txOut.Coins),
		"adjustByInputs: total coins are not equal. diff: %v, tx: %v", incrBalOut.String(), txOut.Coins.String())

}

func TestSendTx(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//ExecTx
	tx := ttx.MakeSendTx(1, et.accOut, et.accIn)
	et.acc2State(et.accIn)
	et.acc2State(et.accOut)
	et.signTx(tx, et.accIn)

	//Bad Balance
	et.accIn.Balance = types.Coins{{"mycoin", 2}}
	et.acc2State(et.accIn)
	res, _, _, _, _ := et.exec(tx, true)
	assert.True(res.IsErr(), "ExecTx/Bad CheckTx: Expected error return from ExecTx, returned: %v", res)

	res, balIn, balInExp, balOut, balOutExp := et.exec(tx, false)
	assert.True(res.IsErr(), "ExecTx/Bad DeliverTx: Expected error return from ExecTx, returned: %v", res)
	assert.False(balIn.IsEqual(balInExp),
		"ExecTx/Bad DeliverTx: balance shouldn't be equal for accIn: got %v, expected: %v", balIn, balInExp)
	assert.False(balOut.IsEqual(balOutExp),
		"ExecTx/Bad DeliverTx: balance shouldn't be equal for accOut: got %v, expected: %v", balOut, balOutExp)

	//Regular CheckTx
	et.reset()
	et.acc2State(et.accIn)
	et.acc2State(et.accOut)
	res, _, _, _, _ = et.exec(tx, true)
	assert.True(res.IsOK(), "ExecTx/Good CheckTx: Expected OK return from ExecTx, Error: %v", res)

	//Regular DeliverTx
	et.reset()
	et.acc2State(et.accIn)
	et.acc2State(et.accOut)
	res, balIn, balInExp, balOut, balOutExp = et.exec(tx, false)
	assert.True(res.IsOK(), "ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res)
	assert.True(balIn.IsEqual(balInExp),
		"ExecTx/good DeliverTx: unexpected change in input balance, got: %v, expected: %v", balIn, balInExp)
	assert.True(balOut.IsEqual(balOutExp),
		"ExecTx/good DeliverTx: unexpected change in output balance, got: %v, expected: %v", balOut, balOutExp)
}

// func TestSendTxIBC(t *testing.T) {
// 	assert := assert.New(t)
// 	et := newExecTest()

// 	//ExecTx
// 	chainID2 := "otherchain"
// 	tx := ttx.MakeSendTx(1, et.accOut, et.accIn)
// 	dstAddress := tx.Outputs[0].Address
// 	tx.Outputs[0].Address = []byte(chainID2 + "/" + string(tx.Outputs[0].Address))
// 	et.acc2State(et.accIn)
// 	et.signTx(tx, et.accIn)

// 	//Regular DeliverTx
// 	et.reset()
// 	et.acc2State(et.accIn)

// 	initBalIn := et.state.GetAccount(et.accIn.Account.Username).Balance

// 	res := ExecTx(et.state, nil, tx, false, nil)

// 	balIn := et.state.GetAccount(et.accIn.Account.Username).Balance
// 	decrBalInExp := tx.Outputs[0].Coins.Plus(types.Coins{tx.Fee}) //expected decrease in balance In
// 	balInExp := initBalIn.Minus(decrBalInExp)

// 	assert.True(res.IsOK(), "ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res)
// 	assert.True(balIn.IsEqual(balInExp),
// 		"ExecTx/good DeliverTx: unexpected change in input balance, got: %v, expected: %v", balIn, balInExp)

// 	packet, err := ibc.GetIBCPacket(et.state, et.chainID, chainID2, 0)
// 	assert.Nil(err)

// 	assert.Equal(packet.SrcChainID, et.chainID)
// 	assert.Equal(packet.DstChainID, chainID2)
// 	assert.Equal(packet.Sequence, uint64(0))
// 	assert.Equal(packet.Type, "coin")

// 	coins, ok := packet.Payload.(ibc.CoinsPayload)
// 	assert.True(ok)
// 	assert.Equal(coins.Coins, tx.Outputs[0].Coins)
// 	assert.EqualValues(coins.Address, dstAddress)
// }

func TestPostTx(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	tx := ttx.MakePostTx(et.accOut.Account.Username, 1, et.accOut)
	signBytes := tx.SignBytes(et.chainID)
	tx.Signature = et.accOut.Sign(signBytes)
	//ExecTx
	et.acc2State(et.accOut)

	initPostSeq := et.state.GetAccount(et.accOut.Account.Username).LastPost

	// Test seq equal to 1
	res := ExecTx(et.state, nil, tx, false, nil)
	assert.True(res.IsOK(), "ExecTx/Good PostTx: Expected OK return from ExecTx, Error: %v", res)
	endPostSeq := et.state.GetAccount(et.accOut.Account.Username).LastPost
	assert.Equal(endPostSeq, initPostSeq + 1)

	// Test seq larger than 1
	tx = ttx.MakePostTx(et.accOut.Account.Username, 2, et.accOut)
	signBytes = tx.SignBytes(et.chainID)
	tx.Signature = et.accOut.Sign(signBytes)
	initPostSeq = et.state.GetAccount(et.accOut.Account.Username).LastPost

	res = ExecTx(et.state, nil, tx, false, nil)
	assert.True(res.IsOK(), "ExecTx/Good PostTx: Expected OK return from ExecTx, Error: %v", res)
	endPostSeq = et.state.GetAccount(et.accOut.Account.Username).LastPost
	assert.Equal(endPostSeq, initPostSeq + 1)

	// First post
	acc := ttx.MakeAcc("username", "secret")
	et.acc2State(acc)
	tx = ttx.MakePostTx(acc.Account.Username, 1, acc)
	signBytes = tx.SignBytes(et.chainID)
	tx.Signature = acc.Sign(signBytes)
	preAcc := et.state.GetAccount(acc.Account.Username)


	assert.True(preAcc != nil)
	res = ExecTx(et.state, nil, tx, false, nil)
	assert.True(res.IsOK(), "ExecTx/Good PostTx: Expected OK return from ExecTx, Error: %v", res)
	endPostSeq = et.state.GetAccount(acc.Account.Username).LastPost
	assert.Equal(1, endPostSeq)

	// Test comment
	tx = ttx.MakePostTx(acc.Account.Username, 2, acc)
	tx.Parent = types.GetPostID(acc.Account.Username, 1)
	signBytes = tx.SignBytes(et.chainID)
	tx.Signature = acc.Sign(signBytes)

	res = ExecTx(et.state, nil, tx, false, nil)
	assert.True(res.IsOK(), "ExecTx/Good PostTx: Expected OK return from ExecTx, Error: %v", res)
	endPostSeq = et.state.GetAccount(acc.Account.Username).LastPost
	assert.Equal(2, endPostSeq)

	// Invalid seq no
	tx = ttx.MakePostTx(et.accOut.Account.Username, 100, et.accOut)
	signBytes = tx.SignBytes(et.chainID)
	tx.Signature = et.accOut.Sign(signBytes)
	initPostSeq = et.state.GetAccount(et.accOut.Account.Username).LastPost

	res = ExecTx(et.state, nil, tx, false, nil)
	assert.Equal(abci.CodeType_BaseInvalidSequence, res.Code, "ExecTx/Bad PostTx: expected error on tx input with bad sequence")
	endPostSeq = et.state.GetAccount(et.accOut.Account.Username).LastPost
	assert.Equal(endPostSeq, initPostSeq)

	// Unsigned post
	tx = ttx.MakePostTx(et.accOut.Account.Username, 3, et.accOut)
	signBytes = tx.SignBytes(et.chainID)
	initPostSeq = et.state.GetAccount(et.accOut.Account.Username).LastPost

	res = ExecTx(et.state, nil, tx, false, nil)
	assert.Equal(abci.ErrBaseInvalidSignature.Code, res.Code, "ExecTx/Bad PostTx: expected error on tx input with bad sequence")
	endPostSeq = et.state.GetAccount(et.accOut.Account.Username).LastPost
	assert.Equal(endPostSeq, initPostSeq)

	// Invalid comment
	tx = ttx.MakePostTx(acc.Account.Username, 3, acc)
	tx.Parent = []byte("parent")
	signBytes = tx.SignBytes(et.chainID)
	tx.Signature = acc.Sign(signBytes)

	res = ExecTx(et.state, nil, tx, false, nil)
	assert.Equal(abci.ErrBaseUnknownAddress.Code, res.Code, "ExecTx/Bad PostTx: expected error on tx input with bad sequence")
	endPostSeq = et.state.GetAccount(acc.Account.Username).LastPost
	assert.Equal(2, endPostSeq)
}

func TestDonateTx(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	pstx := ttx.MakePostTx(et.accOut.Account.Username, 1, et.accOut)
	signBytes := pstx.SignBytes(et.chainID)
	pstx.Signature = et.accOut.Sign(signBytes)
	et.acc2State(et.accOut)
	initOutputBalance := et.state.GetAccount(et.accOut.Account.Username).Balance
	initPostSeq := et.state.GetAccount(et.accOut.Account.Username).LastPost

	// Test seq equal to 1
	res := ExecTx(et.state, nil, pstx, false, nil)
	assert.True(res.IsOK(), "ExecTx/Good PostTx: Expected OK return from ExecTx, Error: %v", res)
	endPostSeq := et.state.GetAccount(et.accOut.Account.Username).LastPost
	assert.Equal(endPostSeq, initPostSeq + 1)

	// Test valid donate
	var testCost int64 = 5
	var testFee int64 = 1
	var initBalance int64 = 7

	dtx := ttx.MakeDonateTx(1, testCost, testFee, types.GetPostID(et.accOut.Account.Username, 1), et.accIn)
	et.acc2State(et.accIn)
	et.acc2State(et.accOut)
	dtxSignBytes := dtx.SignBytes(et.chainID)
	dtx.Input.Signature = et.accIn.Sign(dtxSignBytes)

	// Execute donate
	initOutputBalance = et.state.GetAccount(et.accOut.Account.Username).Balance
	assert.True(initOutputBalance.IsEqual(types.Coins{{"mycoin", initBalance}}),
			"ExecTx/Valid DeliverTx: balance should be equal for accIn: got %v, expected: %v", initOutputBalance, initBalance)
	res = ExecTx(et.state, nil, dtx, true, nil)
	assert.True(res.IsOK(), "ExecTx/Good DonateTx: Expected OK return from ExecTx, Error: %v", res)
	endInputBalance := et.state.GetAccount(et.accIn.Account.Username).Balance
	assert.True(endInputBalance.IsEqual(types.Coins{{"mycoin", initBalance - testCost}}),
			"ExecTx/Valid DeliverTx: balance should be equal for accIn: got %v, expected: %v", endInputBalance, initBalance - testCost)

	endOutputBalance := et.state.GetAccount(et.accOut.Account.Username).Balance
	assert.True(endOutputBalance.IsEqual(types.Coins{{"mycoin", initBalance}}),
			"ExecTx/Valid DeliverTx: balance should be equal for accIn: got %v, expected: %v", endOutputBalance, initBalance)

	et.reset()
	et.acc2State(et.accIn)
	et.acc2State(et.accOut)
	res = ExecTx(et.state, nil, pstx, false, nil)
	res = ExecTx(et.state, nil, dtx, false, nil)
	assert.True(res.IsOK(), "ExecTx/Good DonateTx: Expected OK return from ExecTx, Error: %v", res)
	endInputBalance = et.state.GetAccount(et.accIn.Account.Username).Balance
	assert.True(endInputBalance.IsEqual(types.Coins{{"mycoin", initBalance - testCost}}),
			"ExecTx/Valid DeliverTx: balance should be equal for accIn: got %v, expected: %v", endInputBalance, initBalance - testCost)

	endOutputBalance = et.state.GetAccount(et.accOut.Account.Username).Balance
	assert.True(endOutputBalance.IsEqual(types.Coins{{"mycoin", initBalance + testCost - testFee}}),
			"ExecTx/Valid DeliverTx: balance should be equal for accIn: got %v, expected: %v", endOutputBalance, initBalance)
}

func TestLikeTx(t *testing.T) {
	// set up environment
	assert := assert.New(t)
	et := newExecTest()
	seq := 1
	et.acc2State(et.accOut)
	pstTx := ttx.MakePostTx(et.accOut.Account.Username, seq, et.accOut)
	pstID := types.GetPostID(et.accOut.Account.Username, seq)
	pstSignBytes := pstTx.SignBytes(et.chainID)
	pstTx.Signature = et.accOut.Sign(pstSignBytes)
	res := ExecTx(et.state, nil, pstTx, false, nil)
	assert.True(res.IsOK(), "ExecTx/Good PostTx: Expected OK return from ExecTx, Error: %v", res)

	// Valid Like
	tx1 := ttx.MakeLikeTx(10000, et.accOut, pstID)
	signBytes := tx1.SignBytes(et.chainID)
	tx1.Signature = et.accOut.Sign(signBytes)
	rst := ExecTx(et.state, nil, tx1, false, nil)
	assert.True(rst.IsOK(), "LikeTx error: %v", rst)
	likes := et.state.GetLikesByPostId(pstID)
	assert.Equal(1, len(likes), "Unexpeted Likes array: %v", likes)
	assert.Equal(et.accOut.Account.Username, likes[0].From)
	assert.Equal(pstID, likes[0].To)
	assert.Equal(10000, likes[0].Weight)

	// Invalid post Id
	tx1 = ttx.MakeLikeTx(0, et.accOut, []byte("wrong_pid"))
	signBytes = tx1.SignBytes(et.chainID)
	tx1.Signature = et.accOut.Sign(signBytes)
	rst = ExecTx(et.state, nil, tx1, false, nil)
	assert.Equal(abci.ErrBaseUnknownAddress.Code, rst.Code, "ExecTx/Bad PostTx: expected error on tx input with bad sequence")
	likes = et.state.GetLikesByPostId(pstID)
	assert.Equal(1, len(likes), "Unexpeted Likes array: %v", likes)

	// Valid Dislike
	tx1 = ttx.MakeLikeTx(-10000, et.accOut, pstID)
	signBytes = tx1.SignBytes(et.chainID)
	tx1.Signature = et.accOut.Sign(signBytes)
	rst = ExecTx(et.state, nil, tx1, false, nil)
	assert.True(rst.IsOK(), "LikeTx error: %v", rst)
	likes = et.state.GetLikesByPostId(pstID)
	assert.Equal(1, len(likes), "Unexpeted Likes array: %v", likes)
	assert.Equal(pstID, likes[0].To)
	assert.Equal(-10000, likes[0].Weight)
}
