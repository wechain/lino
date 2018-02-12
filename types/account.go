package types

import (
	"fmt"
	"time"
	"github.com/tendermint/go-crypto"
)


type AccountName []byte
type JsonFormat  string
type AccountList []AccountName

type Account struct {
	Username          AccountName   `json:"username"`            // Primary Key
	PubKey            crypto.PubKey `json:"pub_key"`             // May be nil, if not known
	Metadata          JsonFormat    `json:"metadata"`            // Json format metadata
	LastAccountUpdate time.Time     `json:"last_account_update"` // Account last update
	Created           time.Time     `json:"created"`             // Account creation time
	Mined             bool          `json:"mined"`               // Is mine (Not used now)
	LastTransaction   int           `json:"last_transaction"`    // Transaction sequence
	Balance           Coins         `json:"coins"`               // Account balance
	LastPost          int           `json:"last_post"`           // Post sequence
	FollowerList      AccountList   `json:"follower_list"`       // All followers
	FollowingList     AccountList   `json:"following_list"`      // All account followed by this account
}

func (acc *Account) Copy() *Account {
	if acc == nil {
		return nil
	}
	accCopy := *acc
	return &accCopy
}

func (acc *Account) String() string {
	if acc == nil {
		return "nil-Account"
	}
	return fmt.Sprintf(`Account{username:%v, PubKey:%v, Metadata:%v, LastAccountUpdate:%v
		                         Created:%v, Mined:%v, LastTransaction:%v, Balance:%v, LastPost:%v
		                         FollowerList:%v, FollowingList:%v}`,
		acc.Username, acc.PubKey, acc.Metadata, acc.LastAccountUpdate, acc.Created, acc.Mined,
		acc.LastTransaction, acc.Balance, acc.LastPost, acc.FollowerList, acc.FollowingList)
}

//----------------------------------------

type PrivAccount struct {
	crypto.PrivKey
	Account
}

//----------------------------------------

func GetAccountNameFromString(username string) AccountName {
	return AccountName([]byte(username))
}

type AccountGetter interface {
	GetAccount(username AccountName) *Account
}

type AccountSetter interface {
	SetAccount(username AccountName, acc *Account)
}

type AccountGetterSetter interface {
	GetAccount(addr []byte) *Account
	SetAccount(addr []byte, acc *Account)
}
