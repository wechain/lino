package commands

import (
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
var LikeTxCmd = &cobra.Command{
	Use:   "like",
	Short: "like a post based on weight from -10000 to 10000",
	RunE:  commands.RequireInit(doLikeTx),
}

//nolint
const (
	FlagPostAuthor = "postauthor"
	FlagPostSeq    = "postseq"
	FlagWeight     = "weight"
)

func init() {
	flags := LikeTxCmd.Flags()
	flags.String(FlagPostAuthor, "", "Post author username")
	flags.String(FlagName, "", "Username")
	flags.String(FlagPostSeq, "", "Post sequence of that author")
	flags.Int(FlagWeight, 0, "Like weight")
}

// runDemo is an example of how to make a tx
func doLikeTx(cmd *cobra.Command, args []string) error {
	// load data from json or flags
	tx := new(ttx.LikeTx)
	err := readLikeTxFlags(tx)
	if err != nil {
		return err
	}

	// Wrap and add signer
	like := &ttx.CliLikeTx{
		ChainID: commands.GetChainID(),
		Tx:      tx,
	}
	like.AddSigner(txcmd.GetSigner())
	// Sign if needed and post.  This it the work-horse
	bres, err := txcmd.SignAndPostTx(like)
	if err != nil {
		return err
	}

	// Output result
	return txcmd.OutputTx(bres)
}

func readLikeTxFlags(tx *ttx.LikeTx) error {
	//parse the fee and amounts into coin types
	tx.From = btypes.GetAccountNameFromString(viper.GetString(FlagName))
	poster := viper.GetString(FlagPostAuthor)
	tx.To = btypes.GetPostID(btypes.GetAccountNameFromString(poster), viper.GetInt(FlagPostSeq))
	tx.Weight = viper.GetInt(FlagWeight)
	return nil
}