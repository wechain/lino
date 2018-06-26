package post

import (
	"github.com/lino-network/lino/param"
	"github.com/lino-network/lino/types"
	"github.com/lino-network/lino/x/post/model"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PostManager struct {
	postStorage model.PostStorage `json:"post_storage"`
	paramHolder param.ParamHolder `json:"param_holder"`
}

// create NewPostManager
func NewPostManager(key sdk.StoreKey, holder param.ParamHolder) PostManager {
	return PostManager{
		postStorage: model.NewPostStorage(key),
		paramHolder: holder,
	}
}

func (pm PostManager) GetRedistributionSplitRate(ctx sdk.Context, permLink types.PermLink) (sdk.Rat, sdk.Error) {
	postMeta, err := pm.postStorage.GetPostMeta(ctx, permLink)
	if err != nil {
		return sdk.ZeroRat(), ErrGetRedistributionSplitRate(permLink)
	}
	return postMeta.RedistributionSplitRate, nil
}

func (pm PostManager) GetCreatedTimeAndReward(ctx sdk.Context, permLink types.PermLink) (int64, types.Coin, sdk.Error) {
	postMeta, err := pm.postStorage.GetPostMeta(ctx, permLink)
	if err != nil {
		return 0, types.NewCoinFromInt64(0), ErrGetCreatedTime(permLink)
	}
	return postMeta.CreatedAt, postMeta.TotalReward, nil
}

// check if post exist
func (pm PostManager) DoesPostExist(ctx sdk.Context, permLink types.PermLink) bool {
	return pm.postStorage.DoesPostExist(ctx, permLink)
}

// return root source post
func (pm PostManager) GetSourcePost(
	ctx sdk.Context, permLink types.PermLink) (types.AccountKey, string, sdk.Error) {
	postInfo, err := pm.postStorage.GetPostInfo(ctx, permLink)
	if err != nil {
		return types.AccountKey(""), "", ErrGetRootSourcePost(permLink)
	}

	// check source post's source, that's the root
	if postInfo.SourceAuthor == types.AccountKey("") || postInfo.SourcePostID == "" {
		return types.AccountKey(""), "", nil
	}

	return postInfo.SourceAuthor, postInfo.SourcePostID, nil
}

func (pm PostManager) setRootSourcePost(ctx sdk.Context, postInfo *model.PostInfo) sdk.Error {
	if postInfo.SourceAuthor == types.AccountKey("") || postInfo.SourcePostID == "" {
		return nil
	}
	permLink := types.GetPermLink(postInfo.Author, postInfo.PostID)
	rootAuthor, rootPostID, err :=
		pm.GetSourcePost(ctx, types.GetPermLink(postInfo.SourceAuthor, postInfo.SourcePostID))
	if err != nil {
		return ErrSetRootSourcePost(permLink)
	}
	if rootAuthor != types.AccountKey("") && rootPostID != "" {
		postInfo.SourceAuthor = rootAuthor
		postInfo.SourcePostID = rootPostID
	}
	return nil
}

// create the post
func (pm PostManager) CreatePost(
	ctx sdk.Context, author types.AccountKey, postID string,
	sourceAuthor types.AccountKey, sourcePostID string,
	parentAuthor types.AccountKey, parentPostID string,
	content string, title string, redistributionSplitRate sdk.Rat,
	links []types.IDToURLMapping) sdk.Error {
	postInfo := &model.PostInfo{
		PostID:       postID,
		Title:        title,
		Content:      content,
		Author:       author,
		ParentAuthor: parentAuthor,
		ParentPostID: parentPostID,
		SourceAuthor: sourceAuthor,
		SourcePostID: sourcePostID,
		Links:        links,
	}
	permLink := types.GetPermLink(postInfo.Author, postInfo.PostID)
	if pm.DoesPostExist(ctx, permLink) {
		return ErrPostExist(permLink)
	}
	if err := pm.setRootSourcePost(ctx, postInfo); err != nil {
		return ErrCreatePostSourceInvalid(permLink)
	}
	if err := pm.postStorage.SetPostInfo(ctx, postInfo); err != nil {
		return ErrCreatePost(permLink)
	}
	postMeta := &model.PostMeta{
		CreatedAt:               ctx.BlockHeader().Time,
		LastUpdatedAt:           ctx.BlockHeader().Time,
		LastActivityAt:          ctx.BlockHeader().Time,
		AllowReplies:            true, // Default
		IsDeleted:               false,
		RedistributionSplitRate: redistributionSplitRate.Round(types.PrecisionFactor),
	}
	if err := pm.postStorage.SetPostMeta(ctx, permLink, postMeta); err != nil {
		return ErrCreatePost(permLink)
	}
	return nil
}

func (pm PostManager) UpdatePost(
	ctx sdk.Context, author types.AccountKey, postID, title, content string,
	links []types.IDToURLMapping, redistributionSplitRate sdk.Rat) sdk.Error {
	permLink := types.GetPermLink(author, postID)
	postInfo, err := pm.postStorage.GetPostInfo(ctx, permLink)
	if err != nil {
		return err
	}
	postMeta, err := pm.postStorage.GetPostMeta(ctx, permLink)
	if err != nil {
		return err
	}

	postInfo.Title = title
	postInfo.Content = content
	postInfo.Links = links
	postMeta.RedistributionSplitRate = redistributionSplitRate

	if err := pm.postStorage.SetPostInfo(ctx, postInfo); err != nil {
		return err
	}
	if err := pm.postStorage.SetPostMeta(ctx, permLink, postMeta); err != nil {
		return err
	}
	return nil
}

// add or update like from the user if like exists
func (pm PostManager) AddOrUpdateLikeToPost(
	ctx sdk.Context, permLink types.PermLink, user types.AccountKey, weight int64) sdk.Error {
	postMeta, err := pm.postStorage.GetPostMeta(ctx, permLink)
	if err != nil {
		return ErrAddOrUpdateLikeToPost(permLink)
	}
	like, _ := pm.postStorage.GetPostLike(ctx, permLink, user)
	// Revoke privous
	if like != nil {
		if like.Weight > 0 {
			postMeta.TotalLikeWeight -= like.Weight
		}
		if like.Weight < 0 {
			postMeta.TotalDislikeWeight += like.Weight
		}
		like.Weight = weight
	} else {
		postMeta.TotalLikeCount += 1
		like = &model.Like{Username: user, Weight: weight, CreatedAt: ctx.BlockHeader().Time}
	}
	if like.Weight > 0 {
		postMeta.TotalLikeWeight += like.Weight
	}
	if like.Weight < 0 {
		postMeta.TotalDislikeWeight -= like.Weight
	}
	postMeta.LastActivityAt = ctx.BlockHeader().Time
	if err := pm.postStorage.SetPostLike(ctx, permLink, like); err != nil {
		return ErrAddOrUpdateLikeToPost(permLink)
	}
	if err := pm.postStorage.SetPostMeta(ctx, permLink, postMeta); err != nil {
		return ErrAddOrUpdateLikeToPost(permLink)
	}
	return nil
}

// add or update like from the user if like exists
func (pm PostManager) AddOrUpdateViewToPost(
	ctx sdk.Context, permLink types.PermLink, user types.AccountKey) sdk.Error {
	postMeta, err := pm.postStorage.GetPostMeta(ctx, permLink)
	if err != nil {
		return ErrAddOrUpdateViewToPost(permLink)
	}
	view, _ := pm.postStorage.GetPostView(ctx, permLink, user)
	// Revoke previous
	if view == nil {
		view = &model.View{Username: user}
	}
	postMeta.TotalViewCount += 1
	view.Times += 1
	view.LastViewAt = ctx.BlockHeader().Time
	if err := pm.postStorage.SetPostView(ctx, permLink, view); err != nil {
		return ErrAddOrUpdateViewToPost(permLink)
	}
	if err := pm.postStorage.SetPostMeta(ctx, permLink, postMeta); err != nil {
		return ErrAddOrUpdateViewToPost(permLink)
	}
	return nil
}

// add or update report or upvote from the user if exist
func (pm PostManager) ReportOrUpvoteToPost(
	ctx sdk.Context, permLink types.PermLink, user types.AccountKey, stake types.Coin, isReport bool) sdk.Error {
	postMeta, err := pm.postStorage.GetPostMeta(ctx, permLink)
	if err != nil {
		return ErrAddOrUpdateReportOrUpvoteToPost(permLink)
	}
	postMeta.LastActivityAt = ctx.BlockHeader().Time

	reportOrUpvote, _ := pm.postStorage.GetPostReportOrUpvote(ctx, permLink, user)
	// Revoke privous
	if reportOrUpvote != nil {
		return ErrReportOrUpvoteToPostExist(permLink)
	} else {
		reportOrUpvote =
			&model.ReportOrUpvote{Username: user, Stake: stake, CreatedAt: ctx.BlockHeader().Time}
	}
	if isReport {
		postMeta.TotalReportStake = postMeta.TotalReportStake.Plus(reportOrUpvote.Stake)
		reportOrUpvote.IsReport = true
	} else {
		postMeta.TotalUpvoteStake = postMeta.TotalUpvoteStake.Plus(reportOrUpvote.Stake)
		reportOrUpvote.IsReport = false
	}
	if err := pm.postStorage.SetPostReportOrUpvote(ctx, permLink, reportOrUpvote); err != nil {
		return ErrAddOrUpdateReportOrUpvoteToPost(permLink)
	}
	if err := pm.postStorage.SetPostMeta(ctx, permLink, postMeta); err != nil {
		return ErrAddOrUpdateReportOrUpvoteToPost(permLink)
	}
	return nil
}

// add comment to post comment list
func (pm PostManager) AddComment(
	ctx sdk.Context, permLink types.PermLink, commentUser types.AccountKey, commentPostID string) sdk.Error {
	comment := &model.Comment{
		Author:    commentUser,
		PostID:    commentPostID,
		CreatedAt: ctx.BlockHeader().Time,
	}
	if err := pm.postStorage.SetPostComment(ctx, permLink, comment); err != nil {
		return err
	}

	if err := pm.AddCommentToList(ctx, permLink, commentUser, commentPostID); err != nil {
		return err
	}

	return nil
}

func (pm PostManager) AddCommentToList(ctx sdk.Context, permLink types.PermLink, commentUser types.AccountKey, commentPostID string) sdk.Error {
	commentList, err := pm.postStorage.GetCommentList(ctx, permLink)
	if err != nil {
		return err
	}

	if commentList == nil {
		commentList = &model.CommentList{[]types.PermLink{}}
	}

	commentPermLink := types.GetPermLink(commentUser, commentPostID)
	commentList.CommentList = append(commentList.CommentList, commentPermLink)

	return pm.postStorage.SetCommentList(ctx, permLink, commentList)
}

// add donation to post donation list
func (pm PostManager) AddDonation(
	ctx sdk.Context, permLink types.PermLink, donator types.AccountKey,
	amount types.Coin, donationType types.DonationType) sdk.Error {
	postMeta, err := pm.postStorage.GetPostMeta(ctx, permLink)
	if err != nil {
		return ErrAddDonation(permLink)
	}
	donation := model.Donation{
		Amount:       amount,
		CreatedAt:    ctx.BlockHeader().Time,
		DonationType: donationType,
	}
	donations, _ := pm.postStorage.GetPostDonations(ctx, permLink, donator)
	if donations == nil {
		donations = &model.Donations{Username: donator, DonationList: []model.Donation{}}
	}
	donations.DonationList = append(donations.DonationList, donation)
	if err := pm.postStorage.SetPostDonations(ctx, permLink, donations); err != nil {
		return ErrAddDonation(permLink)
	}
	postMeta.TotalReward = postMeta.TotalReward.Plus(donation.Amount)
	postMeta.TotalDonateCount = postMeta.TotalDonateCount + 1
	if err := pm.postStorage.SetPostMeta(ctx, permLink, postMeta); err != nil {
		return ErrAddDonation(permLink)
	}
	return nil
}

// DeletePost triggered by censorship proposal
func (pm PostManager) DeletePost(ctx sdk.Context, permLink types.PermLink) sdk.Error {
	postMeta, err := pm.postStorage.GetPostMeta(ctx, permLink)
	if err != nil {
		return ErrDeletePost(permLink)
	}
	postMeta.IsDeleted = true
	if err := pm.postStorage.SetPostMeta(ctx, permLink, postMeta); err != nil {
		return ErrAddDonation(permLink)
	}
	postInfo, err := pm.postStorage.GetPostInfo(ctx, permLink)
	if err != nil {
		return ErrDeletePost(permLink)
	}
	postInfo.Title = ""
	postInfo.Content = ""
	postInfo.Links = nil

	if err := pm.postStorage.SetPostInfo(ctx, postInfo); err != nil {
		return ErrDeletePost(permLink)
	}
	return nil
}

func (pm PostManager) IsDeleted(ctx sdk.Context, permLink types.PermLink) (bool, sdk.Error) {
	postMeta, err := pm.postStorage.GetPostMeta(ctx, permLink)
	if err != nil {
		return false, ErrDeletePost(permLink)
	}
	return postMeta.IsDeleted, nil
}

// get penalty score from report and upvote
func (pm PostManager) GetPenaltyScore(ctx sdk.Context, permLink types.PermLink) (sdk.Rat, sdk.Error) {
	sourceAuthor, sourcePostID, err := pm.GetSourcePost(ctx, permLink)
	if err != nil {
		return sdk.ZeroRat(), ErrGetPenaltyScore(permLink)
	}
	if sourceAuthor != types.AccountKey("") && sourcePostID != "" {
		paneltyScore, err := pm.GetPenaltyScore(ctx, types.GetPermLink(sourceAuthor, sourcePostID))
		if err != nil {
			return sdk.ZeroRat(), err
		}
		return paneltyScore, nil
	}
	postMeta, err := pm.postStorage.GetPostMeta(ctx, permLink)
	if err != nil {
		return sdk.ZeroRat(), ErrGetPenaltyScore(permLink)
	}
	if postMeta.TotalReportStake.IsZero() {
		return sdk.ZeroRat(), nil
	}
	if postMeta.TotalUpvoteStake.IsZero() {
		return sdk.OneRat(), nil
	}
	penaltyScore := postMeta.TotalReportStake.ToRat().Quo(postMeta.TotalUpvoteStake.ToRat())
	if penaltyScore.GT(sdk.OneRat()) {
		return sdk.OneRat(), nil
	}
	return penaltyScore, nil
}