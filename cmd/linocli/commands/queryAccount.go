package commands

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/spf13/cobra"

	wire "github.com/tendermint/go-wire"
	lc "github.com/tendermint/light-client"
	lcmd "github.com/tendermint/light-client/commands"
	proofcmd "github.com/tendermint/light-client/commands/proofs"
	"github.com/tendermint/light-client/proofs"

	btypes "github.com/lino-network/lino/types"
	"github.com/lino-network/lino/state"
	ttx "github.com/lino-network/lino/types/tx"
)

var AccountQueryCmd = &cobra.Command{
	Use:   "account [address]",
	Short: "Get details of an account, with proof",
	RunE:  lcmd.RequireInit(doAccountQuery),
}

func init() {
	flags := AccountQueryCmd.Flags()
	flags.String(FlagUsername, "", "Username")
}

func doAccountQuery(cmd *cobra.Command, args []string) error {
	user := btypes.GetAccountNameFromString(viper.GetString(FlagUsername))
	key := state.AccountKey(user)

	acc := new(btypes.Account)
	proof, err := proofcmd.GetAndParseAppProof(key, &acc)
	if lc.IsNoDataErr(err) {
		return errors.Errorf("Account bytes are empty for address %X ", user)
	} else if err != nil {
		return err
	}

	return proofcmd.OutputProof(acc, proof.BlockHeight())
}

// BaseTxPresenter this decodes all basecoin tx
type BaseTxPresenter struct {
	proofs.RawPresenter // this handles MakeKey as hex bytes
}

func (_ BaseTxPresenter) ParseData(raw []byte) (interface{}, error) {
	var tx ttx.TxS
	err := wire.ReadBinaryBytes(raw, &tx)
	return tx, err
}
