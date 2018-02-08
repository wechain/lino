package transaction

import (
	abci "github.com/tendermint/abci/types"
	"github.com/lino-network/lino/types"
	. "github.com/tendermint/tmlibs/common"
)

type TxOutput struct {
	Username types.AccountName `json:"username"` // Hash of the PubKey
	Coins    types.Coins       `json:"coins"`   //
}

// An output destined for another chain may be formatted as `chainID/address`.
// ChainAndAddress returns the chainID prefix and the address.
// If there is no chainID prefix, the first returned value is nil.
// func (txOut TxOutput) ChainAndAddress() ([]byte, []byte, abci.Result) {
// 	var chainPrefix []byte
// 	username := txOut.Username
// 	if len(username) > 20 {
// 		spl := bytes.SplitN(username, []byte("/"), 2)
// 		if len(spl) != 2 {
// 			return nil, nil, abci.ErrBaseInvalidOutput.AppendLog("Invalid address format")
// 		}
// 		chainPrefix = spl[0]
// 		username = spl[1]
// 	}

// 	if len(username) != 20 {
// 		return nil, nil, abci.ErrBaseInvalidOutput.AppendLog("Invalid username length")
// 	}
// 	return chainPrefix, address, abci.OK
// }

func (txOut TxOutput) ValidateBasic() abci.Result {
	// _, _, r := txOut.ChainAndAddress()
	// if r.IsErr() {
	// 	return r
	// }
	if !txOut.Coins.IsValid() {
		return abci.ErrBaseInvalidOutput.AppendLog(Fmt("Invalid coins %v", txOut.Coins))
	}
	if txOut.Coins.IsZero() {
		return abci.ErrBaseInvalidOutput.AppendLog("Coins cannot be zero")
	}
	return abci.OK
}

func (txOut TxOutput) String() string {
	return Fmt("TxOutput{%X,%v}", txOut.Username, txOut.Coins)
}