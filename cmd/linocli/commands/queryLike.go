package commands

import (
	"encoding/hex"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmn "github.com/tendermint/tmlibs/common"
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
	flags.String(FlagAddress, "", "Destination address for the query")
	flags.Int(FlagPostSeq, -1, "Sequence number for post")
	flags.String(FlagPostAuthor, "", "Author for post")
}

func doLikeQuery(cmd *cobra.Command, args []string) error {
	poster, err := hex.DecodeString(cmn.StripHex(viper.GetString(FlagPostAuthor)))
	if err != nil {
		return err
	}

	address, AddErr := hex.DecodeString(cmn.StripHex(viper.GetString(FlagAddress)))
	if AddErr != nil {
		return AddErr
	}

	pKey := btypes.PostID(poster, viper.GetInt(FlagPostSeq))
	lKey := btypes.LikeKey(btypes.LikeID(address, pKey))
	like := new(btypes.Like)

	proof, err := proofcmd.GetAndParseAppProof(lKey, &like)
	if err != nil {
		return err
	}

	return proofcmd.OutputProof(like, proof.BlockHeight())
}