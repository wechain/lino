package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	lcmd "github.com/tendermint/light-client/commands"
	proofcmd "github.com/tendermint/light-client/commands/proofs"

	btypes "github.com/lino-network/lino/types"
)

var LikeQueryCmd = &cobra.Command{
	Use:   "like",
	Short: "Get details of an like, with proof",
	RunE:  lcmd.RequireInit(doLikeQuery),
}

func init() {
	flags := LikeQueryCmd.Flags()
	flags.String(FlagUsername, "", "Username")
	flags.Int(FlagPostSeq, -1, "Sequence number for post")
	flags.String(FlagPostAuthor, "", "Author for post")
}

func doLikeQuery(cmd *cobra.Command, args []string) error {
	poster := btypes.GetAccountNameFromString(viper.GetString(FlagPostAuthor))
	user := btypes.GetAccountNameFromString(viper.GetString(FlagUsername))
	pKey := btypes.GetPostID(poster, viper.GetInt(FlagPostSeq))
	lKey := btypes.LikeKey(btypes.LikeID(user, pKey))
	like := new(btypes.Like)

	proof, err := proofcmd.GetAndParseAppProof(lKey, &like)
	if err != nil {
		return err
	}

	return proofcmd.OutputProof(like, proof.BlockHeight())
}