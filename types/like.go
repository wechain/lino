package types

import (
	"fmt"
	"github.com/tendermint/go-wire"
	"bytes"
	"reflect"
)

type Like struct {
	From   []byte `json:"from"`      // address
	To     []byte     `json:"to"`        // post_id
	Weight int        `json:"weight"`    // weight
}

type LikeId []byte

type LikeSummary struct {
	Likes    []LikeId
}
// const LIKE_SUMMARY_KEY [...]byte("LikeSummaryKey")
const (
	LIKE_SUMMARY_KEY = "LIKE_SUMMARY_KEY"
)

func (lk *Like) String() string {
	return fmt.Sprintf("Like{%v, %v, %v}", lk.From, lk.To, lk.Weight)
}

func LikeID(from []byte, to []byte) LikeId{
	return append(from, to...)
}

func LikeKey(kid []byte) LikeId {
	return append([]byte("like/"), kid...)
}

// TODO(djj): disallow invalid Like in CheckTx, instead of leave it as no-op.

func GetLikesByPostId(store KVStore, pid []byte) []Like {
	summary := readLikeSummary(store)
	var likes []Like
	for _, like_id := range summary.Likes {
		if like := GetLike(store, like_id); bytes.Equal(like.To, pid) {
			likes = append(likes, *like)
		}
	}
	return likes
}

func AddLike(store KVStore, like Like) {
	if !doesPostExist(store, like.To) {
		return
	}
	like_id := insertLikeToDb(store, &like)
	summary := readLikeSummary(store)
	// insert like to db
	if !likeExist(store, &like, summary) {
		// update summary
		summary.Likes = append(summary.Likes, like_id)
		updateSummary(store, summary)
	}
}

func updateSummary(store KVStore, summary *LikeSummary) {
	bytes := wire.BinaryBytes(summary)
	store.Set([]byte(LIKE_SUMMARY_KEY), bytes);
}

func insertLikeToDb(store KVStore, like *Like) LikeId {
	bytes := wire.BinaryBytes(like)
	like_id := LikeID(like.From, like.To)
	store.Set(LikeKey(like_id), bytes)
	return like_id
}

// func likeExist(store KVStore, to_insert Like) bool {
// 	summary := readLikeSummary(store)
// 	return likeExist(store, to_insert, summary)
// }

func likeExist(store KVStore, to_insert *Like, summary *LikeSummary) bool {
	for _, like_id := range summary.Likes {
		like := GetLike(store, like_id)
		if reflect.DeepEqual(like, to_insert) {
			return true
		}
	}
	return false
}

func GetLike(store KVStore, like_id LikeId) *Like {
	data := store.Get(LikeKey(like_id))
	if len(data) == 0 {
		return nil
	}
	var like *Like
	err := wire.ReadBinaryBytes(data, &like)
	if err != nil {
		panic("Calling ReadLike using an invalid like_id")
	}
	return like
}

func readLikeSummary(store KVStore) *LikeSummary {
	data := store.Get([]byte(LIKE_SUMMARY_KEY))
	if len(data) == 0 {
		return &LikeSummary{}
	}
	var summary *LikeSummary
	err := wire.ReadBinaryBytes(data, &summary)
	if err != nil {
		panic("ReadLikeSummary is corrupted.")
	}
	return summary
}

func doesPostExist(store KVStore, pid []byte) bool {
	return GetPost(store, pid) != nil
}