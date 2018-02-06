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
var DonateTxCmd = &cobra.Command{
	Use:   "donate",
	Short: "donate to a post",
	RunE:  commands.RequireInit(doDonateTx),
}


func init() {
	flags := DonateTxCmd.Flags()
	flags.String(FlagPostAuthor, "", "Post author")
	flags.String(FlagPostSeq, "", "Post sequence of that author")
	flags.String(FlagAmount, "", "Coins to send in the format <amt><coin>,<amt><coin>...")
	flags.String(FlagFee, "0mycoin", "Coins for the transaction fee of the format <amt><coin>")
	flags.Int(FlagSequence, -1, "Sequence number for this transaction")
}

// runDemo is an example of how to make a tx
func doDonateTx(cmd *cobra.Command, args []string) error {
	// load data from json or flags
	tx := new(ttx.DonateTx)
	err := readDonateTxFlags(tx)
	if err != nil {
		return err
	}

	// Wrap and add signer
	donate := &ttx.CliDonateTx{
		ChainID: commands.GetChainID(),
		Tx:      tx,
	}
	donate.AddSigner(txcmd.GetSigner())
	// Sign if needed and post.  This it the work-horse
	bres, err := txcmd.SignAndPostTx(donate)
	if err != nil {
		return err
	}

	// Output result
	return txcmd.OutputTx(bres)
}

func readDonateTxFlags(tx *ttx.DonateTx) error {
	poster, err := hex.DecodeString(cmn.StripHex(viper.GetString(FlagPostAuthor)))
	if err != nil {
		return errors.Wrap(err, "Invalid address")
	}
	tx.To = btypes.PostID(poster, viper.GetInt(FlagPostSeq))
	tx.Fee, err = btypes.ParseCoin(viper.GetString(FlagFee))
	if err != nil {
		return err
	}

	amountCoins, err := btypes.ParseCoins(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}
	tx.Input = ttx.TxInput{
		Coins:    amountCoins,
		Sequence: viper.GetInt(FlagSequence),
	}
	return nil
}
