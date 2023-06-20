package tail

import (
	"errors"
	"fmt"
	"os"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	"github.com/juju/ansiterm"
	"github.com/spf13/cobra"

	"github.com/maticnetwork/avail-settlement/pkg/avail"
	"github.com/maticnetwork/avail-settlement/pkg/block"
)

func GetCommand() *cobra.Command {
	var availAddr string
	var offset int64
	cmd := &cobra.Command{
		Use:   "tail",
		Short: "Follow Avail SL blockstream from Avail",
		Run: func(cmd *cobra.Command, args []string) {
			Run(availAddr, offset)
		},
	}
	cmd.Flags().StringVar(&availAddr, "avail-addr", "ws://127.0.0.1:9944/v1/json-rpc", "Avail JSON-RPC URL")
	cmd.Flags().Int64Var(&offset, "offset", 1, "Block offset; defaults to first block")
	return cmd
}

func Run(availAddr string, offset int64) {
	availClient, err := avail.NewClient(availAddr, hclog.NewNullLogger())
	if err != nil {
		panic(err)
	}

	appID, err := avail.QueryAppID(availClient, avail.ApplicationKey)
	if err != nil {
		panic(err)
	}

	callIdx, err := avail.FindCallIndex(availClient)
	if err != nil {
		panic(err)
	}

	availBlkStream := availClient.BlockStream(1)

	tw := ansiterm.NewTabWriter(os.Stdout, 4, 4, 1, ' ', 0)

	for blk := range availBlkStream.Chan() {
		blks, err := block.FromAvail(blk, appID, callIdx, hclog.NewNullLogger())
		if err != nil && !errors.Is(err, block.ErrNoExtrinsicFound) {
			panic(err)
		}

		for _, b := range blks {
			if b.Number() < uint64(offset) {
				continue
			}

			printBlock(tw, b)
		}

		tw.Flush()
	}
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

func printBlock(tw *ansiterm.TabWriter, blk *types.Block) {
	switch {
	case blk.Number() == 0:
		printGenesis(tw, blk)
	case isFraudProofBlock(blk):
		printFraudProofBlock(tw, blk)
	case isBeginDisputeResolutionBlock(blk):
		printBeginDisputeResolutionBlock(tw, blk)
	case isEndDisputeResolutionBlock(blk):
		printSlashBlock(tw, blk)
	default:
		printDefaultBlock(tw, blk)
	}
}

func printGenesis(tw *ansiterm.TabWriter, blk *types.Block) {
	tw.SetForeground(ansiterm.Magenta)
	fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%s\n", blk.Number(), blk.Hash().String(), blk.ParentHash().String(), len(blk.Transactions), "GENESIS")
	tw.Reset()
}

func printDefaultBlock(tw *ansiterm.TabWriter, blk *types.Block) {
	tw.SetForeground(ansiterm.Gray)
	fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%s\n", blk.Number(), blk.Hash().String(), blk.ParentHash().String(), len(blk.Transactions), "NORMAL")
	tw.Reset()
}

func printFraudProofBlock(tw *ansiterm.TabWriter, blk *types.Block) {
	tw.SetForeground(ansiterm.BrightYellow)
	fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%s\n", blk.Number(), blk.Hash().String(), blk.ParentHash().String(), len(blk.Transactions), "FRAUDPROOF")
	tw.Reset()
}

func printBeginDisputeResolutionBlock(tw *ansiterm.TabWriter, blk *types.Block) {
	// TODO: Check if block is forking or not; take it into account when selecting color.
	tw.SetForeground(ansiterm.BrightBlue)
	fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%s\n", blk.Number(), blk.Hash().String(), blk.ParentHash().String(), len(blk.Transactions), "BEGIN DISPUTE RESOLUTION")
	tw.Reset()
}

func printSlashBlock(tw *ansiterm.TabWriter, blk *types.Block) {
	tw.SetForeground(ansiterm.BrightRed)
	fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%s\n", blk.Number(), blk.Hash().String(), blk.ParentHash().String(), len(blk.Transactions), "END DISPUTE RESOLUTION")
	tw.Reset()
}

func isFraudProofBlock(blk *types.Block) bool {
	_, fpExists := block.GetExtraDataFraudProofTarget(blk.Header)
	_, bdrExists := block.GetExtraDataBeginDisputeResolutionTarget(blk.Header)
	return fpExists && bdrExists
}

func isBeginDisputeResolutionBlock(blk *types.Block) bool {
	_, exists := block.GetExtraDataBeginDisputeResolutionTarget(blk.Header)
	return exists
}

func isEndDisputeResolutionBlock(blk *types.Block) bool {
	_, exists := block.GetExtraDataEndDisputeResolutionTarget(blk.Header)
	return exists
}
