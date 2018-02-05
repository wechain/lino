package transaction

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	"github.com/lino-network/lino/types"
	. "github.com/tendermint/tmlibs/common"
	keys "github.com/tendermint/go-crypto/keys"
)

type AppTx struct {
	Gas   int64           `json:"gas"`   // Gas
	Fee   types.Coin            `json:"fee"`   // Fee
	Name  string          `json:"type"`  // Which plugin
	Input TxInput         `json:"input"` // Hmmm do we want coins?
	Data  json.RawMessage `json:"data"`
}

func (_ *AppTx) AssertIsTx()  {}

func (tx *AppTx) SignBytes(chainID string) []byte {
	signBytes := wire.BinaryBytes(chainID)
	sig := tx.Input.Signature
	tx.Input.Signature = crypto.Signature{}
	signBytes = append(signBytes, wire.BinaryBytes(tx)...)
	tx.Input.Signature = sig
	return signBytes
}

func (tx *AppTx) SetSignature(sig crypto.Signature) bool {
	tx.Input.Signature = sig
	return true
}

func (tx *AppTx) String() string {
	return Fmt("AppTx{%v/%v %v %v %X}", tx.Gas, tx.Fee, tx.Name, tx.Input, tx.Data)
}

// ============================================================================
// CliAppTx Application transaction structure for client
type CliAppTx struct {
	ChainID string
	signers []crypto.PubKey
	Tx      *AppTx
}

var _ keys.Signable = &CliAppTx{}

// SignBytes returned the unsigned bytes, needing a signature
func (s *CliAppTx) SignBytes() []byte {
	return s.Tx.SignBytes(s.ChainID)
}

// Sign will add a signature and pubkey.
//
// Depending on the Signable, one may be able to call this multiple times for multisig
// Returns error if called with invalid data or too many times
func (s *CliAppTx) Sign(pubkey crypto.PubKey, sig crypto.Signature) error {
	if len(s.signers) > 0 {
		return errors.New("AppTx already signed")
	}
	s.Tx.SetSignature(sig)
	s.signers = []crypto.PubKey{pubkey}
	return nil
}

// Signers will return the public key(s) that signed if the signature
// is valid, or an error if there is any issue with the signature,
// including if there are no signatures
func (s *CliAppTx) Signers() ([]crypto.PubKey, error) {
	if len(s.signers) == 0 {
		return nil, errors.New("No signatures on AppTx")
	}
	return s.signers, nil
}

// TxBytes returns the transaction data as well as all signatures
// It should return an error if Sign was never called
func (s *CliAppTx) TxBytes() ([]byte, error) {
	// TODO: verify it is signed

	// Code and comment from: basecoin/cmd/basecoin/commands/tx.go
	// Don't you hate having to do this?
	// How many times have I lost an hour over this trick?!
	txBytes := wire.BinaryBytes(TxS{s.Tx})
	return txBytes, nil
}

// TODO: this should really be in the basecoin.types SendTx,
// but that code is too ugly now, needs refactor..
func (a *CliAppTx) ValidateBasic() error {
	if a.ChainID == "" {
		return errors.New("No chain-id specified")
	}
	in := a.Tx.Input
	if len(in.Address) != 20 {
		return errors.Errorf("Invalid input address length: %d", len(in.Address))
	}
	if !in.Coins.IsValid() {
		return errors.Errorf("Invalid input coins %v", in.Coins)
	}
	if in.Coins.IsZero() {
		return errors.New("Input coins cannot be zero")
	}
	if in.Sequence <= 0 {
		return errors.New("Sequence must be greater than 0")
	}
	return nil
}