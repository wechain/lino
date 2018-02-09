package state

import (
	"fmt"
	"time"
	"reflect"
	abci "github.com/tendermint/abci/types"
	"github.com/lino-network/lino/types"
	eyes "github.com/tendermint/merkleeyes/client"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/tmlibs/log"
)

// CONTRACT: State should be quick to copy.
// See CacheWrap().
type State struct {
	chainID    string
	store      types.KVStore
	readCache  map[string][]byte // optional, for caching writes to store
	writeCache *types.KVCache    // optional, for caching writes w/o writing to store
	logger     log.Logger
}

func NewState(store types.KVStore) *State {
	return &State{
		chainID:    "",
		store:      store,
		readCache:  make(map[string][]byte),
		writeCache: nil,
		logger:     log.NewNopLogger(),
	}
}

func (s *State) SetLogger(l log.Logger) {
	s.logger = l
}

func (s *State) SetChainID(chainID string) {
	s.chainID = chainID
	s.store.Set([]byte("base/chain_id"), []byte(chainID))
}

func (s *State) GetChainID() string {
	if s.chainID != "" {
		return s.chainID
	}
	s.chainID = string(s.store.Get([]byte("base/chain_id")))
	return s.chainID
}

func (s *State) Get(key []byte) (value []byte) {
	if s.readCache != nil { //if not a cachewrap
		value, ok := s.readCache[string(key)]
		if ok {
			return value
		}
	}
	return s.store.Get(key)
}

func (s *State) Set(key []byte, value []byte) {
	if s.readCache != nil { //if not a cachewrap
		s.readCache[string(key)] = value
	}
	s.store.Set(key, value)
}


// Account
func AccountKey(username types.AccountName) []byte {
	return append([]byte("account/"), username...)
}

func (s *State) GetAccount(username types.AccountName) *types.Account {
	data := s.Get(AccountKey(username))
	if len(data) == 0 {
		return nil
	}
	var acc *types.Account
	err := wire.ReadBinaryBytes(data, &acc)
	if err != nil {
		panic(fmt.Sprintf("Error reading account %X error: %v",
			data, err.Error()))
	}
	return acc
}

func (s *State) SetAccount(username types.AccountName, acc *types.Account) {
	accBytes := wire.BinaryBytes(acc)
	s.Set(AccountKey(username), accBytes)
}

// Post
func PostKey(pid types.PostID) []byte {
	return append([]byte("post/"), pid...)
}

func (s *State) GetPost(pid types.PostID) *types.Post {
	data := s.Get(PostKey(pid))
	if len(data) == 0 {
		return nil
	}
	var post *types.Post
	err := wire.ReadBinaryBytes(data, &post)
	if err != nil {
		panic(fmt.Sprintf("Error reading Post %X error: %v",
			data, err.Error()))
	}
	return post
}

func (s *State) SetPost(pid types.PostID, post *types.Post) {
	postBytes := wire.BinaryBytes(post)
	s.Set(PostKey(pid), postBytes)
}

func (s *State) PostTxUpdateState(post *types.Post, acc *types.Account, parent *types.Post) {
	if !reflect.DeepEqual(post.Author, acc.Username) {
		panic("post author is different with acc username")
	}
	acc.LastPost += 1
	acc.LastAccountUpdate = time.Now()
	s.SetAccount(acc.Username, acc)
	s.SetPost(types.GetPostID(post.Author, post.Sequence), post)
	if parent != nil {
		if !reflect.DeepEqual(post.Parent, types.GetPostID(parent.Author, parent.Sequence)) {
			panic("post parent doesn't match")
		}
		parent.LastActivity = time.Now()
		parent.Comments = append(parent.Comments, types.GetPostID(post.Author, post.Sequence))
		s.SetPost(post.Parent, post)
	}
}

// Like

func (s *State) GetLikesByPostId(post_id types.PostID) []types.Like {
	return types.GetLikesByPostId(s, post_id);
}

func (s *State) AddLike(like types.Like) {
	types.AddLike(s, like)
}

func (s *State) CacheWrap() *State {
	cache := types.NewKVCache(s)
	return &State{
		chainID:    s.chainID,
		store:      cache,
		readCache:  nil,
		writeCache: cache,
		logger:     s.logger,
	}
}

func (s *State) LikeTxUpdateState(like *types.Like, acc *types.Account, post *types.Post) {
	if !reflect.DeepEqual(like.From, acc.Username) {
		panic("Like Username is different with acc username")
	}
	if !reflect.DeepEqual(like.To, types.GetPostID(post.Author, post.Sequence)) {
		panic("Like target post is invald")
	}
	acc.LastAccountUpdate = time.Now()
	post.LastActivity = time.Now()
	post.Likes = append(post.Likes, like.From)
	s.SetAccount(acc.Username, acc)
	s.SetPost(like.To, post)
	s.AddLike(*like)
}

// Donate
func (s *State) DonateTxUpdateState(post *types.Post, inAcc *types.Account, donate types.Coins, fee types.Coin) {
	inAcc.Balance = inAcc.Balance.Minus(donate)
	inAcc.LastAccountUpdate = time.Now()
	inAcc.LastTransaction += 1
	post.Reward = post.Reward.Plus(donate.Minus(types.Coins{fee}))
	post.LastActivity = time.Now()
	s.SetAccount(inAcc.Username, inAcc)
	s.SetPost(types.GetPostID(post.Author, post.Sequence), post)
}

// Follow

func (s *State) FollowTxUpdateState(inAcc *types.Account, outAcc *types.Account) {
	followingList := inAcc.FollowingList
	if i := UsernameIndexInList(outAcc.Username, followingList); i < 0 {
		inAcc.FollowingList = append(inAcc.FollowingList, outAcc.Username)
	}

	followerList := outAcc.FollowerList
	if i := UsernameIndexInList(inAcc.Username, followerList); i < 0 {
		outAcc.FollowerList = append(outAcc.FollowerList, inAcc.Username)
	}
	s.SetAccount(inAcc.Username, inAcc)
	s.SetAccount(outAcc.Username, outAcc)
}

func (s *State) UnfollowTxUpdateState(inAcc *types.Account, outAcc *types.Account) {
	followingList := inAcc.FollowingList
	if i := UsernameIndexInList(outAcc.Username, followingList); i >= 0 {
		inAcc.FollowingList = append(inAcc.FollowingList[:i], inAcc.FollowingList[i+1:]...)
	}

	followerList := outAcc.FollowerList
	if i := UsernameIndexInList(inAcc.Username, followerList); i >= 0 {
		outAcc.FollowerList = append(outAcc.FollowerList[:i], outAcc.FollowerList[i+1:]...)
	}
	s.SetAccount(inAcc.Username, inAcc)
	s.SetAccount(outAcc.Username, outAcc)
}

func UsernameIndexInList(username types.AccountName, userList types.AccountList) int {
	for i, user := range userList {
		if reflect.DeepEqual(username, user) {
			return i
		}
	}
	return -1
}

// NOTE: errors if s is not from CacheWrap()
func (s *State) CacheSync() {
	s.writeCache.Sync()
}

func (s *State) Commit() abci.Result {
	switch s.store.(type) {
	case *eyes.Client:
		s.readCache = make(map[string][]byte)
		return s.store.(*eyes.Client).CommitSync()
	default:
		return abci.NewError(abci.CodeType_InternalError, "can only use Commit if store is merkleeyes")
	}

}
