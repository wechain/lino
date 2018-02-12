package transaction

import (
	"github.com/pkg/errors"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	. "github.com/tendermint/tmlibs/common"
	"github.com/lino-network/lino/types"
	keys "github.com/tendermint/go-crypto/keys"
)

type RegisterTx struct {
	Username  types.AccountName `json:"username"`   // Hash of the PubKey
	PubKey    crypto.PubKey     `json:"title"`
	Signature crypto.Signature  `json:"signature"`  // Depends on the PubKey type and the whole Tx
}
func (_ *RegisterTx) AssertIsTx()  {}

func (tx *RegisterTx) SignBytes(chainID string) []byte {
	signBytes := wire.BinaryBytes(chainID)
	sig := tx.Signature
	tx.Signature = crypto.Signature{}
	signBytes = append(signBytes, wire.BinaryBytes(tx)...)
	tx.Signature = sig
	return signBytes
}

func (tx *RegisterTx) SetSignature(sig crypto.Signature) bool {
	tx.Signature = sig
	return true
}

func (tx RegisterTx) ValidateBasic() abci.Result {
	if len(tx.Username) <= 3 {
		return abci.ErrBaseInvalidInput.AppendLog("Username length must be greater than 3")
	}
	if tx.PubKey.Empty() {
		return abci.ErrBaseInvalidInput.AppendLog("PubKey empty")
	}
	return abci.OK
}

func (tx *RegisterTx) String() string {
	return Fmt("RegisterTx{%v, %v}", tx.Username, tx.PubKey)
}

// ============================================================================
// CliPostTx Application transaction structure for client
type CliRegisterTx struct {
	ChainID string
	signers []crypto.PubKey
	Tx      *RegisterTx
}

var _ keys.Signable = &CliRegisterTx{}

// SignBytes returned the unsigned bytes, needing a signature
func (r *CliRegisterTx) SignBytes() []byte {
	return r.Tx.SignBytes(r.ChainID)
}

// AddSigner sets address and pubkey info on the tx based on the key that
// will be used for signing
func (r *CliRegisterTx) AddSigner(pk crypto.PubKey) {
	r.Tx.PubKey = pk
	r.signers = append(r.signers, pk)
}

// Sign will add a signature and pubkey.
//
// Depending on the Signable, one may be able to call this multiple times for multisig
// Returns error if called with invalid data or too many times
func (r *CliRegisterTx) Sign(pubkey crypto.PubKey, sig crypto.Signature) error {
	set := r.Tx.SetSignature(sig)
	r.Tx.PubKey = pubkey
	if !set {
		return errors.Errorf("Cannot add signature for address %X", pubkey)
	}
	r.signers = append(r.signers, pubkey)
	return nil
}
// Signers will return the public key(s) that signed if the signature
// is valid, or an error if there is any issue with the signature,
// including if there are no signatures
func (r *CliRegisterTx) Signers() ([]crypto.PubKey, error) {
	if len(r.signers) == 0 {
		return nil, errors.New("No signatures on SendTx")
	}
	return r.signers, nil
}

// TxBytes returns the transaction data as well as all signatures
// It should return an error if Sign was never called
func (r *CliRegisterTx) TxBytes() ([]byte, error) {
	// TODO: verify it is signed

	// Code and comment from: basecoin/cmd/basecoin/commands/tx.go
	// Don't you hate having to do this?
	// How many times have I lost an hour over this trick?!
	txBytes := wire.BinaryBytes(struct {
		Tx `json:"unwrap"`
	}{r.Tx})
	return txBytes, nil
}

// TODO: this should really be in the basecoin.types SendTx,
// but that code is too ugly now, needs refactor..
func (r *CliRegisterTx) ValidateBasic() error {
	if len(r.Tx.Username) <= 3 {
		return errors.New("Username length must be greater than 3")
	}

	return nil
}