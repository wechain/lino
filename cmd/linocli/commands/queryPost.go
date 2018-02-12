package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	lcmd "github.com/tendermint/light-client/commands"
	proofcmd "github.com/tendermint/light-client/commands/proofs"

	btypes "github.com/lino-network/lino/types"
	"github.com/lino-network/lino/state"
)

var PostQueryCmd = &cobra.Command{
	Use:   "post",
	Short: "Get specific post, with proof",
	RunE:  lcmd.RequireInit(doPostQuery),
}

func init() {
	flags := PostQueryCmd.Flags()
	flags.String(FlagPostAuthor, "", "Post author")
	flags.Int(FlagPostSeq, -1, "Sequence number for post")
}

func doPostQuery(cmd *cobra.Command, args []string) error {
	poster := btypes.GetAccountNameFromString(viper.GetString(FlagPostAuthor))

	key := state.PostKey(btypes.GetPostID(poster, viper.GetInt(FlagPostSeq)))
	post := new(btypes.Post)
	proof, err := proofcmd.GetAndParseAppProof(key, &post)
	if err != nil {
		return err
	}

	return proofcmd.OutputProof(post, proof.BlockHeight())
}