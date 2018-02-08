package types

import (
	"fmt"
	"github.com/tendermint/go-wire"
)

type Post struct {
	Title    string      `json:"denom"`
	Content  string      `json:"content"`
	Username AccountName `json:"author"`
	Parent   []byte      `json:"parent"` // non-empty if it is a comment.
}

func (post Post) String() string {
	return fmt.Sprintf("author:%v, title:%v, content:%v",
					   post.Username, post.Title, post.Content)
}

// Post id is computed by the address and sequence.
func PostID(addr []byte, seq int) []byte {
	return append(addr, wire.BinaryBytes(seq)...)
}

func PostKey(pid []byte) []byte {
	return append([]byte("post/"), pid...)
}

func GetPost(store KVStore, pid []byte) *Post {
	data := store.Get(PostKey(pid))
	if len(data) == 0 {
		return nil
	}
	var post *Post
	err := wire.ReadBinaryBytes(data, &post)
	if err != nil {
		panic(fmt.Sprintf("Error reading Post %X error: %v",
			data, err.Error()))
	}
	return post
}

func SetPost(store KVStore, pid []byte, post *Post) {
	postBytes := wire.BinaryBytes(post)
	store.Set(PostKey(pid), postBytes)
}
