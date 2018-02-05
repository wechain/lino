package transaction

import (
	"bytes"

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
	Inputs  []TxInput  `json:"inputs"`
	Outputs []TxOutput `json:"outputs"`
}

func (tx *SendTx) SignBytes(chainID string) []byte {
	signBytes := wire.BinaryBytes(chainID)
	sigz := make([]crypto.Signature, len(tx.Inputs))
	for i := range tx.Inputs {
		sigz[i] = tx.Inputs[i].Signature
		tx.Inputs[i].Signature = crypto.Signature{}
	}
	signBytes = append(signBytes, wire.BinaryBytes(tx)...)
	for i := range tx.Inputs {
		tx.Inputs[i].Signature = sigz[i]
	}
	return signBytes
}

func (tx *SendTx) SetSignature(addr []byte, sig crypto.Signature) bool {
	for i, input := range tx.Inputs {
		if bytes.Equal(input.Address, addr) {
			tx.Inputs[i].Signature = sig
			return true
		}
	}
	return false
}

func (tx *SendTx) String() string {
	return Fmt("SendTx{%v/%v %v->%v}", tx.Gas, tx.Fee, tx.Inputs, tx.Outputs)
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
	addr := pubkey.Address()
	set := s.Tx.SetSignature(addr, sig)
	if !set {
		return errors.Errorf("Cannot add signature for address %X", addr)
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

// AddSigner sets address and pubkey info on the tx based on the key that
// will be used for signing
func (s *CliSendTx) AddSigner(pk crypto.PubKey) {
	// get addr if available
	var addr []byte
	if !pk.Empty() {
		addr = pk.Address()
	}

	// set the send address, and pubkey if needed
	in := s.Tx.Inputs
	in[0].Address = addr
	if in[0].Sequence == 1 {
		in[0].PubKey = pk
	}
}

// TODO: this should really be in the basecoin.types SendTx,
// but that code is too ugly now, needs refactor..
func (s *CliSendTx) ValidateBasic() error {
	if s.ChainID == "" {
		return errors.New("No chain-id specified")
	}
	for _, in := range s.Tx.Inputs {
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
	}
	for _, out := range s.Tx.Outputs {
		// we now allow chain/addr, so it can be more than 20 bytes
		if len(out.Address) < 20 {
			return errors.Errorf("Invalid output address length: %d", len(out.Address))
		}
		if !out.Coins.IsValid() {
			return errors.Errorf("Invalid output coins %v", out.Coins)
		}
		if out.Coins.IsZero() {
			return errors.New("Output coins cannot be zero")
		}
	}

	return nil
}