package post

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	acc "github.com/lino-network/lino/tx/account"
	"github.com/lino-network/lino/types"
	"github.com/stretchr/testify/assert"
)

func TestHandlerCreatePost(t *testing.T) {
	ctx, am, pm, gm := setupTest(t, 1)
	handler := NewHandler(pm, am, gm)

	user := createTestAccount(ctx, am, "user1")

	// test valid post
	postCreateParams := PostCreateParams{
		PostID:       "TestPostID",
		Title:        string(make([]byte, 50)),
		Content:      string(make([]byte, 1000)),
		Author:       user,
		ParentAuthor: "",
		ParentPostID: "",
		SourceAuthor: "",
		SourcePostID: "",
		Links:        []types.IDToURLMapping{},
		RedistributionSplitRate: sdk.ZeroRat,
	}
	msg := NewCreatePostMsg(postCreateParams)
	result := handler(ctx, msg)
	assert.True(t, pm.IsPostExist(ctx, types.GetPostKey(postCreateParams.Author, postCreateParams.PostID)))

	// test invlaid author
	postCreateParams.Author = acc.AccountKey("invalid")
	msg = NewCreatePostMsg(postCreateParams)
	result = handler(ctx, msg)
	assert.Equal(t, result, ErrCreatePostAuthorNotFound(postCreateParams.Author).Result())
}

func TestHandlerCreateComment(t *testing.T) {
	ctx, am, pm, gm := setupTest(t, 1)
	handler := NewHandler(pm, am, gm)

	user, postID := createTestPost(t, ctx, "user", "postID", am, pm, string(user))

	// test comment
	postCreateParams := PostCreateParams{
		PostID:       "comment",
		Title:        string(make([]byte, 50)),
		Content:      string(make([]byte, 1000)),
		Author:       user,
		ParentAuthor: user,
		ParentPostID: postID,
		SourceAuthor: "",
		SourcePostID: "",
		Links:        []types.IDToURLMapping{},
		RedistributionSplitRate: sdk.ZeroRat,
	}
	msg := NewCreatePostMsg(postCreateParams)
	result := handler(ctx, msg)
	assert.Equal(t, result, sdk.Result{})

	// after handler check KVStore
	postInfo := model.PostInfo{
		PostID:       postCreateParams.PostID,
		Title:        postCreateParams.Title,
		Content:      postCreateParams.Content,
		Author:       postCreateParams.Author,
		ParentAuthor: postCreateParams.ParentAuthor,
		ParentPostID: postCreateParams.ParentPostID,
		SourceAuthor: postCreateParams.SourceAuthor,
		SourcePostID: postCreateParams.SourcePostID,
		Links:        postCreateParams.Links,
	}

	postMeta := model.PostMeta{
		Created:                 1,
		LastUpdate:              1,
		LastActivity:            1,
		AllowReplies:            true,
		RedistributionSplitRate: sdk.ZeroRat,
	}

	checkPostKVStore(t, ctx, pm, GetPostKey(user, "comment"), postInfo, postMeta)

	// check parent
	postInfo.PostID = "TestPostID"
	postInfo.ParentAuthor = ""
	postInfo.ParentPostID = ""
	postMeta.Created = 0
	postMeta.LastUpdate = 0
	checkPostKVStore(t, ctx, pm, GetPostKey(user, "TestPostID"), postInfo, postMeta)

	// test invalid parent
	postInfo.PostID = "invalid post"
	postInfo.ParentAuthor = user
	postInfo.ParentPostID = "invalid parent"
	msg = NewCreatePostMsg(postInfo)

	result = handler(ctx, msg)
	assert.Equal(t, result, ErrPostMetaNotFound(GetPostMetaKey(GetPostKey(user, "invalid parent"))).Result())

	// test duplicate comment
	postInfo.Author = user
	postInfo.PostID = "comment"
	postInfo.ParentAuthor = user
	postInfo.ParentPostID = "TestPostID"
	msg = NewCreatePostMsg(postInfo)

	result = handler(ctx, msg)
	assert.Equal(t, result, ErrPostExist().Result())

	// test cycle comment
	postInfo.Author = user
	postInfo.PostID = "newComment"
	postInfo.ParentAuthor = user
	postInfo.ParentPostID = "newComment"
	msg = NewCreatePostMsg(postInfo)

	result = handler(ctx, msg)
	assert.Equal(t, result, ErrPostMetaNotFound(GetPostMetaKey(GetPostKey(user, "newComment"))).Result())
}

func TestHandlerRepost(t *testing.T) {
	ctx, am, pm, gm := setupTest(t, 1)
	handler := NewHandler(pm, am, gm)

	user, postID := createTestPost(t, ctx, "user", "postID", am, pm, string(user))

	// test repost
	postCreateParams := PostCreateParams{
		PostID:       "repost",
		Title:        string(make([]byte, 50)),
		Content:      string(make([]byte, 1000)),
		Author:       user,
		ParentAuthor: "",
		ParentPostID: "",
		SourceAuthor: user,
		SourcePostID: postID,
		Links:        []types.IDToURLMapping{},
		RedistributionSplitRate: sdk.ZeroRat,
	}
	msg := NewCreatePostMsg(postCreateParams)
	result := handler(ctx, msg)
	assert.Equal(t, result, sdk.Result{})

	// after handler check KVStore
	postInfo := model.PostInfo{
		PostID:       postCreateParams.PostID,
		Title:        postCreateParams.Title,
		Content:      postCreateParams.Content,
		Author:       postCreateParams.Author,
		ParentAuthor: postCreateParams.ParentAuthor,
		ParentPostID: postCreateParams.ParentPostID,
		SourceAuthor: postCreateParams.SourceAuthor,
		SourcePostID: postCreateParams.SourcePostID,
		Links:        postCreateParams.Links,
	}

	postMeta := model.PostMeta{
		Created:                 1,
		LastUpdate:              1,
		LastActivity:            1,
		AllowReplies:            true,
		RedistributionSplitRate: sdk.ZeroRat,
	}

	checkPostKVStore(t, ctx, pm, GetPostKey(user, "repost"), postInfo, postMeta)

	// test 2 depth repost
	postCreateParams.PostID = "repost-repost"
	postCreateParams.SourceAuthor = user
	postCreateParams.SourcePostID = "repost"
	msg = NewCreatePostMsg(postCreateParams)
	ctx = ctx.WithBlockHeight(2)
	result = handler(ctx, msg)
	assert.Equal(t, result, sdk.Result{})

	// after handler check KVStore
	// check 2 depth repost
	postInfo.PostID = "repost-repost"
	postMeta = PostMeta{
		Created:      2,
		LastUpdate:   2,
		LastActivity: 2,
		AllowReplies: true,
	}
	postInfo.SourceAuthor = user
	postInfo.SourcePostID = "TestPostID"
	checkPostKVStore(t, ctx, pm, GetPostKey(user, "repost-repost"), postInfo, postMeta)
}

func TestHandlerPostLike(t *testing.T) {
	ctx, am, pm, gm := setupTest(t, 1)
	handler := NewHandler(pm, am, gm)

	user := "username"
	postID := "postID"
	createTestAccount(ctx, lam, user)
	createTestPost(t, ctx, lam, pm, user, postID, sdk.ZeroRat)

	likeMsg := NewLikeMsg(acc.AccountKey(user), 10000, acc.AccountKey(user), postID)
	result := handler(ctx, likeMsg)
	assert.Equal(t, result, sdk.Result{})

	// after handler check KVStore
	postInfo := PostInfo{
		PostID:       postID,
		Title:        string(make([]byte, 50)),
		Content:      string(make([]byte, 1000)),
		Author:       acc.AccountKey(user),
		ParentAuthor: "",
		ParentPostID: "",
		SourceAuthor: "",
		SourcePostID: "",
		Links:        []IDToURLMapping{},
		RedistributionSplitRate: sdk.ZeroRat,
	}
	postMeta := PostMeta{
		Created:         0,
		LastUpdate:      0,
		LastActivity:    0,
		AllowReplies:    true,
		TotalLikeCount:  1,
		TotalLikeWeight: 10000,
	}
	checkPostKVStore(t, ctx, pm, GetPostKey(acc.AccountKey(user), postID), postInfo, postMeta)

	// test update like
	likeMsg = NewLikeMsg(acc.AccountKey(user), -10000, acc.AccountKey(user), postID)
	result = handler(ctx, likeMsg)
	assert.Equal(t, result, sdk.Result{})
	postMeta.TotalLikeWeight = -10000
	checkPostKVStore(t, ctx, pm, GetPostKey(acc.AccountKey(user), postID), postInfo, postMeta)

	// test invalid like target post
	likeMsg = NewLikeMsg(acc.AccountKey(user), -10000, acc.AccountKey(user), "invalid")
	result = handler(ctx, likeMsg)
	assert.Equal(t, result, ErrLikePostDoesntExist().Result())
	checkPostKVStore(t, ctx, pm, GetPostKey(acc.AccountKey(user), postID), postInfo, postMeta)

	// test invalid like username
	likeMsg = NewLikeMsg(acc.AccountKey("invalid"), 10000, acc.AccountKey(user), postID)
	result = handler(ctx, likeMsg)

	assert.Equal(t, result, acc.ErrUsernameNotFound().Result())
	checkPostKVStore(t, ctx, pm, GetPostKey(acc.AccountKey(user), postID), postInfo, postMeta)
}

func TestHandlerPostDonate(t *testing.T) {
	ctx, am, pm, gm := setupTest(t, 1)
	handler := NewHandler(pm, am, gm)

	user1 := "user1"
	user2 := "user2"
	postID := "postID"
	accProxy1 := createTestAccount(ctx, lam, user1)
	accProxy2 := createTestAccount(ctx, lam, user2)
	createTestPost(t, ctx, lam, pm, user1, postID, sdk.ZeroRat)

	donateMsg := NewDonateMsg(acc.AccountKey(user2), types.LNO(sdk.NewRat(100)), acc.AccountKey(user1), postID)
	result := handler(ctx, donateMsg)
	assert.Equal(t, result, sdk.Result{})

	// after handler check KVStore
	postInfo := PostInfo{
		PostID:       postID,
		Title:        string(make([]byte, 50)),
		Content:      string(make([]byte, 1000)),
		Author:       acc.AccountKey(user1),
		ParentAuthor: "",
		ParentPostID: "",
		SourceAuthor: "",
		SourcePostID: "",
		Links:        []IDToURLMapping{},
		RedistributionSplitRate: sdk.ZeroRat,
	}
	postMeta := PostMeta{
		Created:          0,
		LastUpdate:       0,
		LastActivity:     0,
		AllowReplies:     true,
		TotalDonateCount: 1,
		TotalReward:      types.Coin{99 * types.Decimals},
	}

	checkPostKVStore(t, ctx, pm, GetPostKey(acc.AccountKey(user1), postID), postInfo, postMeta)

	acc1Balance, _ := accProxy1.GetBankBalance(ctx)
	acc2Balance, _ := accProxy2.GetBankBalance(ctx)
	assert.Equal(t, true, acc1Balance.IsEqual(types.Coin{222 * types.Decimals}))
	assert.Equal(t, true, acc2Balance.IsEqual(types.Coin{23 * types.Decimals}))
	// test invalid donation target
	donateMsg = NewDonateMsg(acc.AccountKey(user1), types.LNO(sdk.NewRat(100)), acc.AccountKey(user1), "invalid")
	result = handler(ctx, donateMsg)
	assert.Equal(t, result, ErrDonatePostDoesntExist().Result())
	checkPostKVStore(t, ctx, pm, GetPostKey(acc.AccountKey(user1), postID), postInfo, postMeta)

	// test invalid user1name
	donateMsg = NewDonateMsg(acc.AccountKey("invalid"), types.LNO(sdk.NewRat(100)), acc.AccountKey(user1), postID)
	result = handler(ctx, donateMsg)

	assert.Equal(t, result, acc.ErrUsernameNotFound().Result())
	checkPostKVStore(t, ctx, pm, GetPostKey(acc.AccountKey(user1), postID), postInfo, postMeta)

	// test insufficient deposit
	donateMsg = NewDonateMsg(acc.AccountKey(user2), types.LNO(sdk.NewRat(100)), acc.AccountKey(user1), postID)
	result = handler(ctx, donateMsg)

	assert.Equal(t, result, acc.ErrAccountCoinNotEnough().Result())
	checkPostKVStore(t, ctx, pm, GetPostKey(acc.AccountKey(user1), postID), postInfo, postMeta)
}

func TestHandlerRePostDonate(t *testing.T) {
	ctx, am, pm, gm := setupTest(t, 1)
	handler := NewHandler(pm, am, gm)

	user1 := "user1"
	user2 := "user2"
	user3 := "user3"
	postID := "postID"
	accProxy1 := createTestAccount(ctx, lam, user1)
	accProxy2 := createTestAccount(ctx, lam, user2)
	accProxy3 := createTestAccount(ctx, lam, user3)
	createTestPost(t, ctx, lam, pm, user1, postID, sdk.NewRat(15, 100))
	postInfo := PostInfo{
		PostID:       "repost",
		Title:        string(make([]byte, 50)),
		Content:      string(make([]byte, 1000)),
		Author:       acc.AccountKey(user2),
		ParentAuthor: "",
		ParentPostID: "",
		SourceAuthor: acc.AccountKey(user1),
		SourcePostID: postID,
		Links:        []IDToURLMapping{},
		RedistributionSplitRate: sdk.NewRat(15, 100),
	}
	msg := NewCreatePostMsg(postInfo)
	result := handler(ctx, msg)
	assert.Equal(t, result, sdk.Result{})

	donateMsg := NewDonateMsg(acc.AccountKey(user3), types.LNO(sdk.NewRat(100)), acc.AccountKey(user2), "repost")
	result = handler(ctx, donateMsg)
	assert.Equal(t, result, sdk.Result{})

	// after handler check KVStore
	// check repost first
	postMeta := PostMeta{
		Created:          0,
		LastUpdate:       0,
		LastActivity:     0,
		AllowReplies:     true,
		TotalDonateCount: 1,
		TotalReward:      types.Coin{sdk.NewRat(15 * types.Decimals).Mul(sdk.NewRat(99, 100)).Evaluate()},
	}

	checkPostKVStore(t, ctx, pm, GetPostKey(acc.AccountKey(user2), "repost"), postInfo, postMeta)

	// check source post
	postMeta.TotalReward = types.Coin{sdk.NewRat(85 * types.Decimals).Mul(sdk.NewRat(99, 100)).Evaluate()}
	postInfo.Author = acc.AccountKey(user1)
	postInfo.PostID = postID
	postInfo.SourceAuthor = ""
	postInfo.SourcePostID = ""

	checkPostKVStore(t, ctx, pm, GetPostKey(acc.AccountKey(user1), postID), postInfo, postMeta)

	acc1Balance, _ := accProxy1.GetBankBalance(ctx)
	acc2Balance, _ := accProxy2.GetBankBalance(ctx)
	acc3Balance, _ := accProxy3.GetBankBalance(ctx)
	assert.Equal(t, true, acc1Balance.IsEqual(types.Coin{123 * types.Decimals}.
		Plus(types.Coin{sdk.NewRat(85 * types.Decimals).Mul(sdk.NewRat(99, 100)).Evaluate()})))
	assert.Equal(t, true, acc2Balance.IsEqual(types.Coin{123 * types.Decimals}.
		Plus(types.Coin{sdk.NewRat(15 * types.Decimals).Mul(sdk.NewRat(99, 100)).Evaluate()})))
	assert.Equal(t, true, acc3Balance.IsEqual(types.Coin{23 * types.Decimals}))
}
