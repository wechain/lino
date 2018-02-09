package transaction

import (
	"encoding/json"

	"github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/data"
	. "github.com/tendermint/tmlibs/common"
)

/*
Tx (Transaction) is an atomic operation on the ledger state.

Account Types:
 - SendTx         Send coins to address
 - AppTx         Send a msg to a contract that runs in the vm
*/
type Tx interface {
	AssertIsTx()
	SignBytes(chainID string) []byte
}

// Types of Tx implementations
const (
	// Account transactions
	TxTypeSend = byte(0x01)
	TxTypeApp  = byte(0x02)
	TxTypePost  = byte(0x03)
	TxTypeFollow  = byte(0x04)
	TxTypeLike  = byte(0x71)
	TxTypeDonate = byte(0x80)
	TxNameSend = "send"
	TxNameApp  = "app"
	TxNamePost  = "post"
	TxNameDonate = "donate"
	TxNameLike = "like"
	TxNameFollow = "follow"
)

func (_ *SendTx) AssertIsTx() {}
func (_ *PostTx) AssertIsTx()  {}
func (_ *DonateTx) AssertIsTx() {}
func (_ *LikeTx) AssertIsTx()  {}

var txMapper data.Mapper

// register both private key types with go-wire/data (and thus go-wire)
func init() {
	txMapper = data.NewMapper(TxS{}).
		RegisterImplementation(&SendTx{}, TxNameSend, TxTypeSend).
		RegisterImplementation(&AppTx{}, TxNameApp, TxTypeApp).
		RegisterImplementation(&PostTx{}, TxNamePost, TxTypePost).
		RegisterImplementation(&DonateTx{}, TxNameDonate, TxTypeDonate).
		RegisterImplementation(&LikeTx{}, TxNameLike, TxTypeLike).
		RegisterImplementation(&FollowTx{}, TxNameFollow, TxTypeFollow)
}

// TxS add json serialization to Tx
type TxS struct {
	Tx `json:"unwrap"`
}

func (p TxS) MarshalJSON() ([]byte, error) {
	return txMapper.ToJSON(p.Tx)
}

func (p *TxS) UnmarshalJSON(data []byte) (err error) {
	parsed, err := txMapper.FromJSON(data)
	if err == nil {
		p.Tx = parsed.(Tx)
	}
	return
}

func TxID(chainID string, tx Tx) []byte {
	signBytes := tx.SignBytes(chainID)
	return wire.BinaryRipemd160(signBytes)
}

//--------------------------------------------------------------------------------

// Contract: This function is deterministic and completely reversible.
func jsonEscape(str string) string {
	escapedBytes, err := json.Marshal(str)
	if err != nil {
		PanicSanity(Fmt("Error json-escaping a string", str))
	}
	return string(escapedBytes)
}
