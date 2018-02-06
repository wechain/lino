package transaction

// Helper functions for testing

import (
	"github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/lino-network/lino/types"
)

// Creates a PrivAccount from secret.
// The amount is not set.
func PrivAccountFromSecret(secret string) types.PrivAccount {
	privKey :=
		crypto.GenPrivKeyEd25519FromSecret([]byte(secret)).Wrap()
	privAccount := types.PrivAccount{
		PrivKey: privKey,
		Account: types.Account{
			PubKey: privKey.PubKey(),
		},
	}
	return privAccount
}

// Make `num` random accounts
func RandAccounts(num int, minAmount int64, maxAmount int64) []types.PrivAccount {
	privAccs := make([]types.PrivAccount, num)
	for i := 0; i < num; i++ {

		balance := minAmount
		if maxAmount > minAmount {
			balance += cmn.RandInt64() % (maxAmount - minAmount)
		}

		privKey := crypto.GenPrivKeyEd25519().Wrap()
		pubKey := privKey.PubKey()
		privAccs[i] = types.PrivAccount{
			PrivKey: privKey,
			Account: types.Account{
				PubKey:  pubKey,
				Balance: types.Coins{types.Coin{"", balance}},
			},
		}
	}

	return privAccs
}

/////////////////////////////////////////////////////////////////

//func MakeAccs(secrets ...string) (accs []PrivAccount) {
//	for _, secret := range secrets {
//		privAcc := PrivAccountFromSecret(secret)
//		privAcc.Account.Balance = Coins{{"mycoin", 7}}
//		accs = append(accs, privAcc)
//	}
//	return
//}

func MakeAcc(secret string) types.PrivAccount {
	privAcc := PrivAccountFromSecret(secret)
	privAcc.Account.Balance = types.Coins{{"mycoin", 7}}
	return privAcc
}

func Accs2TxInputs(seq int, accs ...types.PrivAccount) []TxInput {
	var txs []TxInput
	for _, acc := range accs {
		tx := NewTxInput(
			acc.Account.PubKey,
			types.Coins{{"mycoin", 5}},
			seq)
		txs = append(txs, tx)
	}
	return txs
}

//turn a list of accounts into basic list of transaction outputs
func Accs2TxOutputs(accs ...types.PrivAccount) []TxOutput {
	var txs []TxOutput
	for _, acc := range accs {
		tx := TxOutput{
			acc.Account.PubKey.Address(),
			types.Coins{{"mycoin", 4}}}
		txs = append(txs, tx)
	}
	return txs
}

func MakeSendTx(seq int, accOut types.PrivAccount, accsIn ...types.PrivAccount) *SendTx {
	tx := &SendTx{
		Gas:     0,
		Fee:     types.Coin{"mycoin", 1},
		Inputs:  Accs2TxInputs(seq, accsIn...),
		Outputs: Accs2TxOutputs(accOut),
	}

	return tx
}

func MakePostTx(seq int, accOut types.PrivAccount) *PostTx {
	if seq > 1 {
		return &PostTx{
			Address:  accOut.PubKey.Address(),
			Title:    "Title",
			Content:  "Content",
			Sequence: seq,
		}	
	} else {
		return &PostTx{
			Address:  accOut.PubKey.Address(),
			Title:    "Title",
			Content:  "Content",
			Sequence: seq,
			PubKey: accOut.PubKey,
		}
	}
}

func MakeDonateTx(seq int, cost int64, fee int64, to []byte, accsIn types.PrivAccount) *DonateTx {
	tx := &DonateTx{
		Fee:   types.Coin{"mycoin", fee},
		Input: NewTxInput(
			   	   accsIn.Account.PubKey,
			       types.Coins{{"mycoin", cost}},
			       seq),
		To:    to,
	}
	return tx
}

func MakeLikeTx(weight int, from types.PrivAccount, post_id []byte, is_first_time bool) *LikeTx {
	if is_first_time {
		return &LikeTx{
			From:  from.PubKey.Address(),
			To: post_id,
			Weight: weight,
			PubKey: from.PubKey,
		}
	} else {
		return &LikeTx{
			From:  from.PubKey.Address(),
			To: post_id,
			Weight: weight,
		}
	}
}

func SignTx(chainID string, tx *SendTx, accs ...types.PrivAccount) {
	signBytes := tx.SignBytes(chainID)
	for i, _ := range tx.Inputs {
		tx.Inputs[i].Signature = accs[i].Sign(signBytes)
	}
}
