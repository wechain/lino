package vote

import (
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/lino-network/lino/param"
	"github.com/lino-network/lino/types"
	"github.com/lino-network/lino/x/global"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	acc "github.com/lino-network/lino/x/account"
	abci "github.com/tendermint/abci/types"
	dbm "github.com/tendermint/tmlibs/db"
)

// Construct some global addrs and txs for tests.
var (
	TestAccountKVStoreKey = sdk.NewKVStoreKey("account")
	TestVoteKVStoreKey    = sdk.NewKVStoreKey("vote")
	TestGlobalKVStoreKey  = sdk.NewKVStoreKey("global")
	TestParamKVStoreKey   = sdk.NewKVStoreKey("param")
)

func InitGlobalManager(ctx sdk.Context, gm global.GlobalManager) error {
	return gm.InitGlobalManager(ctx, types.NewCoinFromInt64(10000*types.Decimals))
}

func setupTest(t *testing.T, height int64) (sdk.Context,
	acc.AccountManager, VoteManager, global.GlobalManager) {
	ctx := getContext(height)
	ph := param.NewParamHolder(TestParamKVStoreKey)
	ph.InitParam(ctx)
	accManager := acc.NewAccountManager(TestAccountKVStoreKey, ph)
	voteManager := NewVoteManager(TestVoteKVStoreKey, ph)
	globalManager := global.NewGlobalManager(TestGlobalKVStoreKey, ph)

	cdc := globalManager.WireCodec()
	cdc.RegisterInterface((*types.Event)(nil), nil)
	cdc.RegisterConcrete(acc.ReturnCoinEvent{}, "1", nil)

	err := InitGlobalManager(ctx, globalManager)
	assert.Nil(t, err)
	return ctx, accManager, voteManager, globalManager
}

func getContext(height int64) sdk.Context {
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(TestAccountKVStoreKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(TestVoteKVStoreKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(TestGlobalKVStoreKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(TestParamKVStoreKey, sdk.StoreTypeIAVL, db)

	ms.LoadLatestVersion()

	return sdk.NewContext(ms, abci.Header{Height: height}, false, nil, log.NewNopLogger())
}

// helper function to create an account for testing purpose
func createTestAccount(ctx sdk.Context, am acc.AccountManager, username string, initCoin types.Coin) types.AccountKey {
	priv := crypto.GenPrivKeyEd25519()
	am.CreateAccount(ctx, "referrer", types.AccountKey(username),
		priv.PubKey(), priv.Generate(0).PubKey(), priv.Generate(1).PubKey(), priv.Generate(2).PubKey(), initCoin)
	return types.AccountKey(username)
}

func coinToString(coin types.Coin) string {
	return strconv.FormatInt(coin.ToInt64()/types.Decimals, 10)
}