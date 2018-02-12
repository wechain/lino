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
//nolint
const (
	FlagFollowing = "following"
	FlagIsFollow    = "isfollow"
)

// FollowTxCmd is CLI command to send tokens between basecoin accounts
var FollowTxCmd = &cobra.Command{
	Use:   "follow",
	Short: "follow or unfollow",
	RunE:  commands.RequireInit(doFollowTx),
}

func init() {
	flags := FollowTxCmd.Flags()
	flags.String(FlagName, "", "Follower")
	flags.String(FlagFollowing, "", "Following")
	flags.Bool(FlagIsFollow, true, "Is follow or unfollow")
}

// runDemo is an example of how to make a tx
func doFollowTx(cmd *cobra.Command, args []string) error {
	// load data from json or flags
	tx := new(ttx.FollowTx)
	err := readFollowTxFlags(tx)
	if err != nil {
		return err
	}

	// Wrap and add signer
	follow := &ttx.CliFollowTx{
		ChainID: commands.GetChainID(),
		Tx:      tx,
	}
	follow.AddSigner(txcmd.GetSigner())
	// Sign if needed and post.  This it the work-horse
	bres, err := txcmd.SignAndPostTx(follow)
	if err != nil {
		return err
	}

	// Output result
	return txcmd.OutputTx(bres)
}

func readFollowTxFlags(tx *ttx.FollowTx) error {
	//parse the fee and amounts into coin types
	tx.Follower = btypes.GetAccountNameFromString(viper.GetString(FlagName))
	tx.Following = btypes.GetAccountNameFromString(viper.GetString(FlagFollowing))
	tx.IsFollow = viper.GetBool(FlagIsFollow)
	return nil
}
