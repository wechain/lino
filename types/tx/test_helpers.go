package transaction

// Helper functions for testing

import (
	"github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/lino-network/lino/types"
)

// Creates a PrivAccount from secret.
// The amount is not set.
func PrivAccountFromSecret(username string, secret string) types.PrivAccount {
	privKey :=
		crypto.GenPrivKeyEd25519FromSecret([]byte(secret)).Wrap()
	privAccount := types.PrivAccount{
		PrivKey: privKey,
		Account: types.Account{
			Username: types.AccountName([]byte(username)),
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
				Username: types.AccountName(pubKey.Address()),
				PubKey:   pubKey,
				Balance:  types.Coins{types.Coin{"", balance}},
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

func MakeAcc(username string, secret string) types.PrivAccount {
	privAcc := PrivAccountFromSecret(username, secret)
	privAcc.Account.Balance = types.Coins{{"mycoin", 7}}
	return privAcc
}

func Acc2TxInput(seq int, acc types.PrivAccount) TxInput {
	return NewTxInput(
				acc.Account.Username,
				types.Coins{{"mycoin", 5}},
				seq)
}

//turn a list of accounts into basic list of transaction outputs
func Acc2TxOutput(acc types.PrivAccount) TxOutput {
	return TxOutput{
			acc.Account.Username,
			types.Coins{{"mycoin", 4}}}
}

func MakeSendTx(seq int, accOut types.PrivAccount, accIn types.PrivAccount) *SendTx {
	tx := &SendTx{
		Gas:     0,
		Fee:     types.Coin{"mycoin", 1},
		Input:   Acc2TxInput(seq, accIn),
		Output:  Acc2TxOutput(accOut),
	}

	return tx
}

func MakePostTx(username types.AccountName, seq int, accOut types.PrivAccount) *PostTx {
	return &PostTx{
			Author:   username,
			Title:    "Title",
			Content:  "Content",
			Sequence: seq,
		}
}

func MakeDonateTx(seq int, cost int64, fee int64, to types.PostID, accsIn types.PrivAccount) *DonateTx {
	tx := &DonateTx{
		Fee:   types.Coin{"mycoin", fee},
		Input: NewTxInput(
			   	   accsIn.Account.Username,
			       types.Coins{{"mycoin", cost}},
			       seq),
		To:    to,
	}
	return tx
}

func MakeLikeTx(weight int, from types.PrivAccount, post_id types.PostID) *LikeTx {
	return &LikeTx{
		From:  from.Account.Username,
		To: post_id,
		Weight: weight,
	}
}

func SignTx(chainID string, tx *SendTx, acc types.PrivAccount) {
	signBytes := tx.SignBytes(chainID)
	tx.Input.Signature = acc.Sign(signBytes)
}
