package transaction

import (
	"github.com/pkg/errors"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	"github.com/lino-network/lino/types"
	. "github.com/tendermint/tmlibs/common"
	keys "github.com/tendermint/go-crypto/keys"
)

type SendTx struct {
	Gas     int64      `json:"gas"` // Gas
	Fee     types.Coin `json:"fee"` // Fee
	Input   TxInput    `json:"input"`
	Output  TxOutput   `json:"output"`
}

func (tx *SendTx) SignBytes(chainID string) []byte {
	signBytes := wire.BinaryBytes(chainID)
	sig := tx.Input.Signature
	tx.Input.Signature = crypto.Signature{}
	signBytes = append(signBytes, wire.BinaryBytes(tx)...)
	tx.Input.Signature = sig
	return signBytes
}

func (tx *SendTx) SetSignature(sig crypto.Signature) bool {
	tx.Input.Signature = sig
	return true
}

func (tx *SendTx) String() string {
	return Fmt("SendTx{%v/%v %v->%v}", tx.Gas, tx.Fee, tx.Input, tx.Output)
}

// ============================================================================
// CliSendTx Application transaction structure for client
type CliSendTx struct {
	ChainID string
	signers []crypto.PubKey
	Tx      *SendTx
}

var _ keys.Signable = &CliSendTx{}

// SignBytes returned the unsigned bytes, needing a signature
func (s *CliSendTx) SignBytes() []byte {
	return s.Tx.SignBytes(s.ChainID)
}

// Sign will add a signature and pubkey.
//
// Depending on the Signable, one may be able to call this multiple times for multisig
// Returns error if called with invalid data or too many times
func (s *CliSendTx) Sign(pubkey crypto.PubKey, sig crypto.Signature) error {
	set := s.Tx.SetSignature(sig)
	if !set {
		return errors.Errorf("Cannot add signature for address %X", pubkey)
	}
	s.signers = append(s.signers, pubkey)
	return nil
}

// Signers will return the public key(s) that signed if the signature
// is valid, or an error if there is any issue with the signature,
// including if there are no signatures
func (s *CliSendTx) Signers() ([]crypto.PubKey, error) {
	if len(s.signers) == 0 {
		return nil, errors.New("No signatures on SendTx")
	}
	return s.signers, nil
}
// AddSigner sets address and pubkey info on the tx based on the key that
// will be used for signing

func (s *CliSendTx) AddSigner(pk crypto.PubKey) {
	s.signers = append(s.signers, pk)
}

// TxBytes returns the transaction data as well as all signatures
// It should return an error if Sign was never called
func (s *CliSendTx) TxBytes() ([]byte, error) {
	// TODO: verify it is signed

	// Code and comment from: basecoin/cmd/basecoin/commands/tx.go
	// Don't you hate having to do this?
	// How many times have I lost an hour over this trick?!
	txBytes := wire.BinaryBytes(struct {
		Tx `json:"unwrap"`
	}{s.Tx})
	return txBytes, nil
}

// TODO: this should really be in the basecoin.types SendTx,
// but that code is too ugly now, needs refactor..
func (s *CliSendTx) ValidateBasic() error {
	if s.Tx.Input.Coins.IsValid() {
		return errors.Errorf("Invalid input coins %v", s.Tx.Input.Coins)
	}
	if s.Tx.Input.Coins.IsZero() {
		return errors.New("Input coins cannot be zero")
	}
	if s.Tx.Input.Sequence <= 0 {
		return errors.New("Sequence must be greater than 0")
	}
	if s.Tx.Output.Coins.IsValid() {
		return errors.Errorf("Invalid output coins %v", s.Tx.Output.Coins)
	}
	if s.Tx.Output.Coins.IsZero() {
		return errors.New("Output coins cannot be zero")
	}

	return nil
}