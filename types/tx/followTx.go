package transaction

import (
	"reflect"
	"github.com/pkg/errors"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	. "github.com/tendermint/tmlibs/common"
	keys "github.com/tendermint/go-crypto/keys"
	"github.com/lino-network/lino/types"
)

//-----------------------------------------------------------------------------
type FollowTx struct {
	Follower  types.AccountName `json:"follower"`      // follower
	Following types.AccountName `json:"following"`     // following
	IsFollow  bool              `json:"is_follow"`     // Differenciate follow and unfollow
	Signature crypto.Signature  `json:"signature"`
}

func (_ *FollowTx) AssertIsTx()  {}

func (tx *FollowTx) SignBytes(chainID string) []byte {
	signBytes := wire.BinaryBytes(chainID)
	sig := tx.Signature
	tx.Signature = crypto.Signature{}
	signBytes = append(signBytes, wire.BinaryBytes(tx)...)
	tx.Signature = sig
	return signBytes
}

func (tx *FollowTx) SetSignature(sig crypto.Signature) bool {
	tx.Signature = sig
	return true
}

func (tx FollowTx) ValidateBasic() abci.Result {
	if len(tx.Follower) == 0 {
		return abci.ErrBaseInvalidInput.AppendLog("Invalid Follower Username")
	}

	if len(tx.Following) == 0 {
		return abci.ErrBaseInvalidInput.AppendLog("Invalid Following Username")
	}

	if reflect.DeepEqual(tx.Following, tx.Follower) {
		return abci.ErrBaseInvalidInput.AppendLog("Can't follow your self")
	}
	return abci.OK
}

func (tx *FollowTx) String() string {
	return Fmt("FollowTx{is follow:%v, from %v to %v, %v}",
		tx.IsFollow, tx.Follower, tx.Following, tx.Signature)
}

// ============================================================================
// CliLikeTx Application transaction structure for client

type CliFollowTx struct {
	ChainID string
	signers []crypto.PubKey
	Tx      *FollowTx
}

var _ keys.Signable = &CliFollowTx{}

// SignBytes returned the unsigned bytes, needing a signature
func (ftx *CliFollowTx) SignBytes() []byte {
	return ftx.Tx.SignBytes(ftx.ChainID)
}

// AddSigner sets address and pubkey info on the tx based on the key that
// will be used for signing
func (ftx *CliFollowTx) AddSigner(pk crypto.PubKey) {
	ftx.signers = append(ftx.signers, pk)
}

// Sign will add a signature and pubkey.
//
// Depending on the Signable, one may be able to call this multiple times for multisig
// Returns error if called with invalid data or too many times
func (ftx *CliFollowTx) Sign(pubkey crypto.PubKey, sig crypto.Signature) error {
	addr := pubkey.Address()
	set := ftx.Tx.SetSignature(sig)
	if !set {
		return errors.Errorf("Cannot add signature for address %X", addr)
	}
	return nil
}
// Signers will return the public key(s) that signed if the signature
// is valid, or an error if there is any issue with the signature,
// including if there are no signatures
func (ftx *CliFollowTx) Signers() ([]crypto.PubKey, error) {
	if len(ftx.signers) == 0 {
		return nil, errors.New("No signatures on SendTx")
	}
	return ftx.signers, nil
}

// TxBytes returns the transaction data as well as all signatures
// It should return an error if Sign was never called
func (ftx *CliFollowTx) TxBytes() ([]byte, error) {
	// TODO: verify it is signed

	// Code and comment from: basecoin/cmd/basecoin/commands/tx.go
	// Don't you hate having to do this?
	// How many times have I lost an hour over this trick?!
	txBytes := wire.BinaryBytes(struct {
		Tx `json:"unwrap"`
	}{ftx.Tx})
	return txBytes, nil
}

// TODO: this should really be in the basecoin.types SendTx,
// but that code is too ugly now, needs refactor..
func (ftx *CliFollowTx) ValidateBasic() error {
	if ftx.ChainID == "" {
		return errors.New("No chain-id specified")
	}

	if reflect.DeepEqual(ftx.Tx.Following, ftx.Tx.Follower) {
		return abci.ErrBaseInvalidInput.AppendLog("Can't follow your self")
	}
	return nil
}