package account

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/lino-network/lino/types"
	"github.com/tendermint/go-crypto"
)

// linoaccount proxy encapsulates all basic struct
type AccountProxy struct {
	username       AccountKey      `json:"username"`
	writeInfoFlag  bool            `json:"write_info_flag"`
	writeBankFlag  bool            `json:"write_bank_flag"`
	writeMetaFlag  bool            `json:"write_meta_flag"`
	accountManager *AccountManager `json:"account_manager"`
	accountInfo    *AccountInfo    `json:"account_info"`
	accountBank    *AccountBank    `json:"account_bank"`
	accountMeta    *AccountMeta    `json:"account_meta"`
}

func RegisterWireLinoAccount(cdc *wire.Codec) {
	// Register crypto.[PubKey] types.
	wire.RegisterCrypto(cdc)
}

// NewLinoAccount return the account pointer
func NewAccountProxy(username AccountKey, accManager *AccountManager) *AccountProxy {
	return &AccountProxy{
		username:       username,
		accountManager: accManager,
	}
}

// check if account exist
func (acc *AccountProxy) IsAccountExist(ctx sdk.Context) bool {
	if err := acc.checkAccountInfo(ctx); err != nil {
		return false
	}
	return true
}

// Implements types.AccountManager.
func (acc *AccountProxy) CreateAccount(ctx sdk.Context, accKey AccountKey, pubkey crypto.PubKey, accBank *AccountBank) sdk.Error {
	if acc.IsAccountExist(ctx) {
		return ErrAccountCreateFail(accKey)
	}
	acc.writeInfoFlag = true
	acc.accountInfo = &AccountInfo{
		Username: accKey,
		Created:  types.Height(ctx.BlockHeight()),
		PostKey:  pubkey,
		OwnerKey: pubkey,
		Address:  pubkey.Address(),
	}

	acc.writeBankFlag = true
	accBank.Username = accKey
	acc.accountBank = accBank

	acc.writeMetaFlag = true
	acc.accountMeta = &AccountMeta{
		LastActivity:   types.Height(ctx.BlockHeight()),
		ActivityBurden: types.DefaultActivityBurden,
	}
	return nil
}

func (acc *AccountProxy) AddCoin(ctx sdk.Context, coin types.Coin) (err sdk.Error) {
	if err := acc.checkAccountBank(ctx); err != nil {
		return err
	}
	acc.accountBank.Balance = acc.accountBank.Balance.Plus(coin)
	acc.writeBankFlag = true
	return nil
}

func (acc *AccountProxy) MinusCoin(ctx sdk.Context, coin types.Coin) (err sdk.Error) {
	if err := acc.checkAccountBank(ctx); err != nil {
		return err
	}

	if !acc.accountBank.Balance.IsGTE(coin) {
		return ErrAccountCoinNotEnough()
	}

	acc.accountBank.Balance = acc.accountBank.Balance.Minus(coin)

	acc.writeBankFlag = true
	return nil
}

func (acc *AccountProxy) GetUsername(ctx sdk.Context) AccountKey {
	return acc.username
}

func (acc *AccountProxy) GetBankAddress(ctx sdk.Context) (sdk.Address, sdk.Error) {
	if err := acc.checkAccountInfo(ctx); err != nil {
		return nil, err
	}
	return acc.accountInfo.Address, nil
}

func (acc *AccountProxy) GetOwnerKey(ctx sdk.Context) (*crypto.PubKey, sdk.Error) {
	if err := acc.checkAccountInfo(ctx); err != nil {
		return nil, err
	}
	return &acc.accountInfo.OwnerKey, nil
}

func (acc *AccountProxy) GetPostKey(ctx sdk.Context) (*crypto.PubKey, sdk.Error) {
	if err := acc.checkAccountInfo(ctx); err != nil {
		return nil, err
	}
	return &acc.accountInfo.PostKey, nil
}

func (acc *AccountProxy) GetBankBalance(ctx sdk.Context) (types.Coin, sdk.Error) {
	if err := acc.checkAccountBank(ctx); err != nil {
		return types.Coin{}, err
	}
	return acc.accountBank.Balance, nil
}

func (acc *AccountProxy) GetSequence(ctx sdk.Context) (int64, sdk.Error) {
	if err := acc.checkAccountMeta(ctx); err != nil {
		return 0, err
	}
	return acc.accountMeta.Sequence, nil
}

func (acc *AccountProxy) GetCreated(ctx sdk.Context) (types.Height, sdk.Error) {
	if err := acc.checkAccountInfo(ctx); err != nil {
		return types.Height(0), err
	}
	return acc.accountInfo.Created, nil
}

func (acc *AccountProxy) GetLastActivity(ctx sdk.Context) (types.Height, sdk.Error) {
	if err := acc.checkAccountMeta(ctx); err != nil {
		return types.Height(0), err
	}
	return acc.accountMeta.LastActivity, nil
}

func (acc *AccountProxy) IncreaseSequenceByOne(ctx sdk.Context) sdk.Error {
	if err := acc.checkAccountMeta(ctx); err != nil {
		return err
	}
	acc.accountMeta.Sequence += 1
	acc.writeMetaFlag = true
	return nil
}

func (acc *AccountProxy) GetActivityBurden(ctx sdk.Context) (int64, sdk.Error) {
	if err := acc.checkAccountMeta(ctx); err != nil {
		return 0, err
	}
	return acc.accountMeta.ActivityBurden, nil
}

func (acc *AccountProxy) UpdateLastActivity(ctx sdk.Context) sdk.Error {
	if err := acc.checkAccountMeta(ctx); err != nil {
		return err
	}
	acc.writeMetaFlag = true
	acc.accountMeta.LastActivity = types.Height(ctx.BlockHeight())
	return nil
}

func (acc *AccountProxy) IsMyFollower(ctx sdk.Context, follower AccountKey) bool {
	return acc.accountManager.IsMyFollower(ctx, acc.username, follower)
}

func (acc *AccountProxy) IsMyFollowing(ctx sdk.Context, followee AccountKey) bool {
	return acc.accountManager.IsMyFollowing(ctx, acc.username, followee)
}

func (acc *AccountProxy) SetFollower(ctx sdk.Context, follower AccountKey) sdk.Error {
	if err := acc.UpdateLastActivity(ctx); err != nil {
		return err
	}

	if acc.IsMyFollower(ctx, follower) {
		return nil
	}
	meta := FollowerMeta{
		CreatedAt:    types.Height(ctx.BlockHeight()),
		FollowerName: follower,
	}
	acc.accountManager.SetFollowerMeta(ctx, acc.username, meta)
	return nil
}

func (acc *AccountProxy) SetFollowing(ctx sdk.Context, followee AccountKey) sdk.Error {
	if err := acc.UpdateLastActivity(ctx); err != nil {
		return err
	}

	if acc.IsMyFollowing(ctx, followee) {
		return nil
	}
	meta := FollowingMeta{
		CreatedAt:    types.Height(ctx.BlockHeight()),
		FolloweeName: followee,
	}
	acc.accountManager.SetFollowingMeta(ctx, acc.username, meta)
	return nil
}

func (acc *AccountProxy) RemoveFollower(ctx sdk.Context, follower AccountKey) sdk.Error {
	if !acc.accountManager.IsMyFollower(ctx, acc.username, follower) {
		return nil
	}
	acc.accountManager.RemoveFollowerMeta(ctx, acc.username, follower)
	return nil
}

func (acc *AccountProxy) RemoveFollowing(ctx sdk.Context, followee AccountKey) sdk.Error {
	if !acc.accountManager.IsMyFollowing(ctx, acc.username, followee) {
		return nil
	}
	acc.accountManager.RemoveFollowingMeta(ctx, acc.username, followee)
	return nil
}

func (acc *AccountProxy) Apply(ctx sdk.Context) sdk.Error {
	if acc.writeInfoFlag {
		if err := acc.accountManager.SetInfo(ctx, acc.username, acc.accountInfo); err != nil {
			return err
		}
	}
	if acc.writeBankFlag {
		if err := acc.checkAccountInfo(ctx); err != nil {
			return err
		}
		if err := acc.accountManager.SetBankFromAddress(ctx, acc.accountInfo.Address, acc.accountBank); err != nil {
			return err
		}
	}
	if acc.writeMetaFlag {
		if err := acc.accountManager.SetMeta(ctx, acc.username, acc.accountMeta); err != nil {
			return err
		}
	}

	acc.clear()
	return nil
}

func (acc *AccountProxy) clear() {
	acc.writeInfoFlag = false
	acc.writeBankFlag = false
	acc.writeMetaFlag = false
	acc.accountInfo = nil
	acc.accountBank = nil
	acc.accountMeta = nil
}

func (acc *AccountProxy) checkAccountInfo(ctx sdk.Context) (err sdk.Error) {
	if acc.accountInfo == nil {
		acc.accountInfo, err = acc.accountManager.GetInfo(ctx, acc.username)
	}
	return err
}

func (acc *AccountProxy) checkAccountBank(ctx sdk.Context) (err sdk.Error) {
	if err = acc.checkAccountInfo(ctx); err != nil {
		return err
	}
	if acc.accountBank == nil {
		acc.accountBank, err = acc.accountManager.GetBankFromAddress(ctx, acc.accountInfo.Address)
	}
	return err
}

func (acc *AccountProxy) checkAccountMeta(ctx sdk.Context) (err sdk.Error) {
	if acc.accountMeta == nil {
		acc.accountMeta, err = acc.accountManager.GetMeta(ctx, acc.username)
	}
	return err
}