package transaction

import (
	"github.com/pkg/errors"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/data"
	. "github.com/tendermint/tmlibs/common"
	keys "github.com/tendermint/go-crypto/keys"
)

//-----------------------------------------------------------------------------
// feature-like
type LikeTx struct {
	From      data.Bytes       `json:"from"`      // address
	To        []byte           `json:"to"`        // post_id
	Weight    int              `json:"weight"`    // like weight from -10000 to 10000
	Signature crypto.Signature `json:"signature"` // Depends on the PubKey type and the whole Tx
	PubKey    crypto.PubKey    `json:"pub_key"`   // Is present iff Sequence == 0
}

func (tx *LikeTx) SignBytes(chainID string) []byte {
	signBytes := wire.BinaryBytes(chainID)
	sig := tx.Signature
	tx.Signature = crypto.Signature{}
	signBytes = append(signBytes, wire.BinaryBytes(tx)...)
	tx.Signature = sig
	return signBytes
}

func (tx *LikeTx) SetSignature(sig crypto.Signature) bool {
	tx.Signature = sig
	return true
}

func (tx LikeTx) ValidateBasic() abci.Result {
	if len(tx.From) != 20 {
		return abci.ErrBaseInvalidInput.AppendLog("Invalid address length")
	}
	if tx.Weight < -10000 || tx.Weight > 10000 {
		return abci.ErrBaseInvalidInput.AppendLog("Invalid weight")
	}
	return abci.OK
}

func (tx *LikeTx) String() string {
	return Fmt("LikeTx{ set %v from %v to %v. (%v, %v) }",
		tx.Weight, tx.From, tx.To, tx.Signature, tx.PubKey)
}

// ============================================================================
// CliLikeTx Application transaction structure for client

type CliLikeTx struct {
	ChainID string
	signers []crypto.PubKey
	Tx      *LikeTx
}

var _ keys.Signable = &CliLikeTx{}

// SignBytes returned the unsigned bytes, needing a signature
func (ltx *CliLikeTx) SignBytes() []byte {
	return ltx.Tx.SignBytes(ltx.ChainID)
}

// AddSigner sets address and pubkey info on the tx based on the key that
// will be used for signing
func (ltx *CliLikeTx) AddSigner(pk crypto.PubKey) {
	ltx.Tx.From = pk.Address()
	ltx.Tx.PubKey = pk
}

// Sign will add a signature and pubkey.
//
// Depending on the Signable, one may be able to call this multiple times for multisig
// Returns error if called with invalid data or too many times
func (ltx *CliLikeTx) Sign(pubkey crypto.PubKey, sig crypto.Signature) error {
	addr := pubkey.Address()
	set := ltx.Tx.SetSignature(sig)
	if !set {
		return errors.Errorf("Cannot add signature for address %X", addr)
	}
	return nil
}
// Signers will return the public key(s) that signed if the signature
// is valid, or an error if there is any issue with the signature,
// including if there are no signatures
func (ltx *CliLikeTx) Signers() ([]crypto.PubKey, error) {
	if len(ltx.signers) == 0 {
		return nil, errors.New("No signatures on SendTx")
	}
	return ltx.signers, nil
}

// TxBytes returns the transaction data as well as all signatures
// It should return an error if Sign was never called
func (ltx *CliLikeTx) TxBytes() ([]byte, error) {
	// TODO: verify it is signed

	// Code and comment from: basecoin/cmd/basecoin/commands/tx.go
	// Don't you hate having to do this?
	// How many times have I lost an hour over this trick?!
	txBytes := wire.BinaryBytes(struct {
		Tx `json:"unwrap"`
	}{ltx.Tx})
	return txBytes, nil
}

// TODO: this should really be in the basecoin.types SendTx,
// but that code is too ugly now, needs refactor..
func (ltx *CliLikeTx) ValidateBasic() error {
	if ltx.ChainID == "" {
		return errors.New("No chain-id specified")
	}
	if len(ltx.Tx.From) != 20 {
		return errors.Errorf("Invalid address length: %d", len(ltx.Tx.From))
	}

	return nil
}