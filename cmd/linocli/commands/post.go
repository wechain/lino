package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/light-client/commands"
	txcmd "github.com/tendermint/light-client/commands/txs"

	btypes "github.com/lino-network/lino/types"
	ttx "github.com/lino-network/lino/types/tx"
)

//-------------------------
// SendTx

// SendTxCmd is CLI command to send tokens between basecoin accounts
var PostTxCmd = &cobra.Command{
	Use:   "post",
	Short: "send a short post",
	RunE:  commands.RequireInit(doPostTx),
}

//nolint
const (
	FlagTitle        = "title"
	FlagContent      = "content"
	FlagParentAuthor = "parentauthor"
	FlagParentSeq    = "parentseq"
	FlagSourceAuthor = "sourceauthor"
	FlagSourceSeq    = "sourceseq"
)

func init() {
	flags := PostTxCmd.Flags()
	flags.String(FlagName, "", "Username")
	flags.String(FlagParentAuthor, "", "Parent post author")
	flags.Int(FlagParentSeq, -1, "Parent post sequence")
	flags.String(FlagSourceAuthor, "", "Source post author")
	flags.Int(FlagSourceSeq, -1, "Source post sequence")
	flags.String(FlagTitle, "", "Post title")
	flags.String(FlagContent, "", "Post content")
	flags.Int(FlagPostSeq, -1, "Sequence number for this post")
}

// runDemo is an example of how to make a tx
func doPostTx(cmd *cobra.Command, args []string) error {
	// load data from json or flags
	tx := new(ttx.PostTx)
	err := readPostTxFlags(tx)
	if err != nil {
		return err
	}

	// Wrap and add signer
	post := &ttx.CliPostTx{
		ChainID: commands.GetChainID(),
		Tx:      tx,
	}
	fmt.Println(post)
	post.AddSigner(txcmd.GetSigner())
	// Sign if needed and post.  This it the work-horse
	bres, err := txcmd.SignAndPostTx(post)
	if err != nil {
		return err
	}

	// Output result
	return txcmd.OutputTx(bres)
}

func readPostTxFlags(tx *ttx.PostTx) error {
	username := viper.GetString(FlagName)
	tx.Author = btypes.GetAccountNameFromString(username)
	parentAuthor := viper.GetString(FlagParentAuthor)
	if len(parentAuthor) > 0 {
		tx.Parent = btypes.GetPostID(btypes.GetAccountNameFromString(parentAuthor), viper.GetInt(FlagParentSeq))
	}
	sourceAuthor := viper.GetString(FlagSourceAuthor)
	if len(sourceAuthor) > 0 {
		tx.Source = btypes.GetPostID(btypes.GetAccountNameFromString(sourceAuthor), viper.GetInt(FlagSourceSeq))
	}
	tx.Title = viper.GetString(FlagTitle)
	tx.Content = viper.GetString(FlagContent)
	tx.Sequence = viper.GetInt(FlagPostSeq)
	return nil
}
