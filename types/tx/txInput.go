package transaction

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	"github.com/lino-network/lino/types"
	. "github.com/tendermint/tmlibs/common"
)

type TxInput struct {
	Username  types.AccountName `json:"address"`   // Unique user identifier
	Coins     types.Coins       `json:"coins"`     //
	Sequence  int               `json:"sequence"`  // Must be 1 greater than the last committed TxInput
	Signature crypto.Signature  `json:"signature"` // Depends on the PubKey type and the whole Tx
}

func (txIn TxInput) ValidateBasic() abci.Result {
	if !txIn.Coins.IsValid() {
		return abci.ErrBaseInvalidInput.AppendLog(Fmt("Invalid coins %v", txIn.Coins))
	}
	if txIn.Coins.IsZero() {
		return abci.ErrBaseInvalidInput.AppendLog("Coins cannot be zero")
	}
	if txIn.Sequence <= 0 {
		return abci.ErrBaseInvalidInput.AppendLog("Sequence must be greater than 0")
	}
	return abci.OK
}

func (txIn TxInput) String() string {
	return Fmt("TxInput{%X,%v,%v,%v,%v}", txIn.Username, txIn.Coins, txIn.Sequence, txIn.Signature)
}

func NewTxInput(username types.AccountName, coins types.Coins, sequence int) TxInput {
	input := TxInput{
		Username:  username,
		Coins:    coins,
		Sequence: sequence,
	}
	return input
}