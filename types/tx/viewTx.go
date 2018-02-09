package transaction

import (
	"github.com/pkg/errors"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	. "github.com/tendermint/tmlibs/common"
	keys "github.com/tendermint/go-crypto/keys"
	"github.com/lino-network/lino/types"
)

//-----------------------------------------------------------------------------
type ViewTx struct {
	From      types.AccountName `json:"from"`      // address
	To        types.PostID      `json:"to"`        // post_id
	Signature crypto.Signature  `json:"signature"` // Depends on the PubKey type and the whole Tx
}

func (_ *ViewTx) AssertIsTx() {}

func (tx *ViewTx) SignBytes(chainID string) []byte {
	signBytes := wire.BinaryBytes(chainID)
	sig := tx.Signature
	tx.Signature = crypto.Signature{}
	signBytes = append(signBytes, wire.BinaryBytes(tx)...)
	tx.Signature = sig
	return signBytes
}

func (tx *ViewTx) SetSignature(sig crypto.Signature) bool {
	tx.Signature = sig
	return true
}

func (tx ViewTx) ValidateBasic() abci.Result {
	if len(tx.From) < 6 {
		return abci.ErrBaseInvalidInput.AppendLog("Invalid Usernamew")
	}
	return abci.OK
}

func (tx *ViewTx) String() string {
	return Fmt("ViewTx{ from %v to %v. (%v) }",
		tx.From, tx.To, tx.Signature)
}

// ============================================================================
// CliViewTx Application transaction structure for client

type CliViewTx struct {
	ChainID string
	signers []crypto.PubKey
	Tx      *LikeTx
}

var _ keys.Signable = &CliViewTx{}

// SignBytes returned the unsigned bytes, needing a signature
func (vtx *CliViewTx) SignBytes() []byte {
	return vtx.Tx.SignBytes(vtx.ChainID)
}

// AddSigner sets address and pubkey info on the tx based on the key that
// will be used for signing
func (vtx *CliViewTx) AddSigner(pk crypto.PubKey) {
	vtx.signers = append(vtx.signers, pk)
}

// Sign will add a signature and pubkey.
//
// Depending on the Signable, one may be able to call this multiple times for multisig
// Returns error if called with invalid data or too many times
func (vtx *CliViewTx) Sign(pubkey crypto.PubKey, sig crypto.Signature) error {
	addr := pubkey.Address()
	set := vtx.Tx.SetSignature(sig)
	if !set {
		return errors.Errorf("Cannot add signature for address %X", addr)
	}
	return nil
}
// Signers will return the public key(s) that signed if the signature
// is valid, or an error if there is any issue with the signature,
// including if there are no signatures
func (vtx *CliViewTx) Signers() ([]crypto.PubKey, error) {
	if len(vtx.signers) == 0 {
		return nil, errors.New("No signatures on SendTx")
	}
	return vtx.signers, nil
}

// TxBytes returns the transaction data as well as all signatures
// It should return an error if Sign was never called
func (vtx *CliViewTx) TxBytes() ([]byte, error) {
	// TODO: verify it is signed

	// Code and comment from: basecoin/cmd/basecoin/commands/tx.go
	// Don't you hate having to do this?
	// How many times have I lost an hour over this trick?!
	txBytes := wire.BinaryBytes(struct {
		Tx `json:"unwrap"`
	}{vtx.Tx})
	return txBytes, nil
}

// TODO: this should really be in the basecoin.types SendTx,
// but that code is too ugly now, needs refactor..
func (vtx *CliViewTx) ValidateBasic() error {
	if vtx.ChainID == "" {
		return errors.New("No chain-id specified")
	}

	return nil
}