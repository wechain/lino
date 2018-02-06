package commands

import (
	"encoding/hex"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmn "github.com/tendermint/tmlibs/common"
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
	FlagPostAuthor = "postAuthor"
	FlagPostSeq    = "postSeq"
	FlagWeight     = "weight"
)

func init() {
	flags := LikeTxCmd.Flags()
	flags.String(FlagPostAuthor, "", "Post author")
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
	poster, posterErr := hex.DecodeString(cmn.StripHex(viper.GetString(FlagPostAuthor)))
	if posterErr != nil {
		return errors.Wrap(posterErr, "Invalid address")
	}
	tx.To = btypes.PostID(poster, viper.GetInt(FlagPostSeq))
	tx.Weight = viper.GetInt(FlagWeight)
	return nil
}
