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
var DonateTxCmd = &cobra.Command{
	Use:   "donate",
	Short: "donate to a post",
	RunE:  commands.RequireInit(doDonateTx),
}


func init() {
	flags := DonateTxCmd.Flags()
	flags.String(FlagName, "", "Username")
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
	poster := viper.GetString(FlagPostAuthor)
	tx.To = btypes.GetPostID(btypes.GetAccountNameFromString(poster), viper.GetInt(FlagPostSeq))
	fee, err := btypes.ParseCoin(viper.GetString(FlagFee))
	if err != nil {
		return err
	}
	tx.Fee = fee

	amountCoins, err := btypes.ParseCoins(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}
	tx.Input = ttx.TxInput{
		Username: btypes.GetAccountNameFromString(viper.GetString(FlagName)),
		Coins:    amountCoins,
		Sequence: viper.GetInt(FlagSequence),
	}
	return nil
}
