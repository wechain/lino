package transaction

import (
	keys "github.com/tendermint/go-crypto/keys"
	"github.com/pkg/errors"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	"github.com/lino-network/lino/types"
	. "github.com/tendermint/tmlibs/common"
)

type DonateTx struct {
	Input     TxInput           `json:"inputTx"` // Hmmm do we want coins?
    To        types.PostID      `json:"to"`      //post_id
    Fee       types.Coin        `json:"fee"`
}

func (tx *DonateTx) SignBytes(chainID string) []byte {
	signBytes := wire.BinaryBytes(chainID)
	sig := tx.Input.Signature
	tx.Input.Signature = crypto.Signature{}
	signBytes = append(signBytes, wire.BinaryBytes(tx)...)
	tx.Input.Signature = sig
	return signBytes
}

func (tx *DonateTx) SetSignature(sig crypto.Signature) bool {
    tx.Input.Signature = sig
    return true
}

func (tx DonateTx) ValidateBasic() abci.Result {
	if err := tx.Input.ValidateBasic(); !err.IsErr() {
		return err
	}
	return abci.OK
}

func (tx *DonateTx) String() string {
        return Fmt("DonateTx{%v -> %v, %v}", tx.Input, tx.To, tx.Fee)
}

// ============================================================================
// CliDonateTx Application transaction structure for client
type CliDonateTx struct {
	ChainID string
	signers []crypto.PubKey
	Tx      *DonateTx
}

var _ keys.Signable = &CliSendTx{}

// SignBytes returned the unsigned bytes, needing a signature
func (s *CliDonateTx) SignBytes() []byte {
	return s.Tx.SignBytes(s.ChainID)
}

// Sign will add a signature and pubkey.
//
// Depending on the Signable, one may be able to call this multiple times for multisig
// Returns error if called with invalid data or too many times
func (s *CliDonateTx) Sign(pubkey crypto.PubKey, sig crypto.Signature) error {
	set := s.Tx.SetSignature(sig)
	if !set {
		return errors.Errorf("Cannot add signature for address %X", pubkey.Address())
	}
	s.signers = append(s.signers, pubkey)
	return nil
}

// Signers will return the public key(s) that signed if the signature
// is valid, or an error if there is any issue with the signature,
// including if there are no signatures
func (s *CliDonateTx) Signers() ([]crypto.PubKey, error) {
	if len(s.signers) == 0 {
		return nil, errors.New("No signatures on SendTx")
	}
	return s.signers, nil
}

// TxBytes returns the transaction data as well as all signatures
// It should return an error if Sign was never called
func (s *CliDonateTx) TxBytes() ([]byte, error) {
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
func (s *CliDonateTx) AddSigner(pk crypto.PubKey) {
	// get addr if available
	s.signers = append(s.signers, pk)
}

// TODO: this should really be in the basecoin.types SendTx,
// but that code is too ugly now, needs refactor..
func (s *CliDonateTx) ValidateBasic() error {
	if s.ChainID == "" {
		return errors.New("No chain-id specified")
	}
	in := s.Tx.Input
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