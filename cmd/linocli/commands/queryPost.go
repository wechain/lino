package commands

import (
	"encoding/hex"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmn "github.com/tendermint/tmlibs/common"
	lcmd "github.com/tendermint/light-client/commands"
	proofcmd "github.com/tendermint/light-client/commands/proofs"

	btypes "github.com/lino-network/lino/types"
)

//nolint
const (
	FlagAddress       = "address"
)

var PostQueryCmd = &cobra.Command{
	Use:   "post",
	Short: "Get specific post, with proof",
	RunE:  lcmd.RequireInit(doPostQuery),
}

func init() {
	flags := PostQueryCmd.Flags()
	flags.String(FlagAddress, "", "Destination address for the query")
	flags.Int(FlagSequence, -1, "Sequence number for post")
}

func doPostQuery(cmd *cobra.Command, args []string) error {
	poster, err := hex.DecodeString(cmn.StripHex(FlagAddress))
	if err != nil {
		return errors.Wrap(err, "Invalid address")
	}
	if err != nil {
		return err
	}

	key := btypes.PostKey(btypes.PostID(poster, viper.GetInt(FlagSequence)))
	post := new(btypes.Post)
	proof, err := proofcmd.GetAndParseAppProof(key, &post)
	if err != nil {
		return err
	}

	return proofcmd.OutputProof(post, proof.BlockHeight())
}