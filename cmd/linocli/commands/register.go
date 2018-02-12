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
var RegisterTxCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a User",
	RunE:  doRegisterTx,
}

//nolint
const (
	FlagName = "name"
	FlagUsername = "username"
)

// runDemo is an example of how to make a tx
func doRegisterTx(cmd *cobra.Command, args []string) error {
	// load data from json or flags
	tx := new(ttx.RegisterTx)
	tx.Username = btypes.GetAccountNameFromString(viper.GetString(FlagName))

	// Wrap and add signer
	register := &ttx.CliRegisterTx{
		ChainID: commands.GetChainID(),
		Tx:      tx,
	}
	register.AddSigner(txcmd.GetSigner())
	// Sign if needed and post.  This it the work-horse
	bres, err := txcmd.SignAndPostTx(register)
	if err != nil {
		return err
	}

	// Output result
	return txcmd.OutputTx(bres)
}
