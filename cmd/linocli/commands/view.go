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
// RegisterTx

// RegisterTxCmd is CLI command to send tokens between basecoin accounts
var ViewTxCmd = &cobra.Command{
	Use:   "view",
	Short: "record the view behavior",
	RunE:  commands.RequireInit(doViewTx),
}

func init() {
	flags := ViewTxCmd.Flags()
	flags.String(FlagPostAuthor, "", "Post author username")
	flags.String(FlagName, "", "Username")
	flags.String(FlagPostSeq, "", "Post sequence of that author")
}

// runDemo is an example of how to make a tx
func doViewTx(cmd *cobra.Command, args []string) error {
	// load data from json or flags
	tx := new(ttx.ViewTx)
	err := readViewTxFlags(tx)
	if err != nil {
		return err
	}

	// Wrap and add signer
	view := &ttx.CliViewTx{
		ChainID: commands.GetChainID(),
		Tx:      tx,
	}
	view.AddSigner(txcmd.GetSigner())
	// Sign if needed and post.  This it the work-horse
	bres, err := txcmd.SignAndPostTx(view)
	if err != nil {
		return err
	}

	// Output result
	return txcmd.OutputTx(bres)
}

func readViewTxFlags(tx *ttx.ViewTx) error {
	//parse the fee and amounts into coin types
	tx.From = btypes.GetAccountNameFromString(viper.GetString(FlagName))
	poster := btypes.GetAccountNameFromString(viper.GetString(FlagPostAuthor))
	tx.To = btypes.GetPostID(poster, viper.GetInt(FlagPostSeq))
	return nil
}
