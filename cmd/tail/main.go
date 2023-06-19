package tail

import (
	"errors"
	"fmt"
	"os"

	edge_types "github.com/0xPolygon/polygon-edge/types"
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

	table := ansiterm.NewTabWriter(os.Stdout, 4, 4, 1, ' ', 0)

	for blk := range availBlkStream.Chan() {
		blks, err := block.FromAvail(blk, appID, callIdx, hclog.NewNullLogger())
		if err != nil && !errors.Is(err, block.ErrNoExtrinsicFound) {
			panic(err)
		}

		for _, b := range blks {
			if b.Number() < uint64(offset) {
				continue
			}

			printBlock(table, b)
		}

		table.Flush()
	}
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

func printBlock(table *ansiterm.TabWriter, blk *edge_types.Block) {
	switch {
	case blk.Number() == 0:
		printGenesis(table, blk)
	case isFraudProofBlock(blk):
		printFraudProofBlock(table, blk)
	case isBeginDisputeResolutionBlock(blk):
		printBeginDisputeResolutionBlock(table, blk)
	case isEndDisputeResolutionBlock(blk):
		printSlashBlock(table, blk)
	default:
		printDefaultBlock(table, blk)
	}
}

//       nbr hash parent nTxs description
// BLK:  %d  %s   %s     %d   %s

func printGenesis(table *ansiterm.TabWriter, blk *edge_types.Block) {
	table.SetForeground(ansiterm.Magenta)
	fmt.Fprintf(table, "%d\t%s\t%s\t%d\t%s\n", blk.Number(), blk.Hash().String(), blk.ParentHash().String(), len(blk.Transactions), "GENESIS")
	table.Reset()
}

func printDefaultBlock(table *ansiterm.TabWriter, blk *edge_types.Block) {
	table.SetForeground(ansiterm.Gray)
	fmt.Fprintf(table, "%d\t%s\t%s\t%d\t%s\n", blk.Number(), blk.Hash().String(), blk.ParentHash().String(), len(blk.Transactions), "DEFAULT")
	table.Reset()
}

func printFraudProofBlock(table *ansiterm.TabWriter, blk *edge_types.Block) {
	table.SetForeground(ansiterm.BrightYellow)
	fmt.Fprintf(table, "%d\t%s\t%s\t%d\t%s\n", blk.Number(), blk.Hash().String(), blk.ParentHash().String(), len(blk.Transactions), "FRAUDPROOF")
	table.Reset()
}

func printBeginDisputeResolutionBlock(table *ansiterm.TabWriter, blk *edge_types.Block) {
	// TODO: Check if block is forking or not; take it into account when selecting color.
	table.SetForeground(ansiterm.BrightBlue)
	fmt.Fprintf(table, "%d\t%s\t%s\t%d\t%s\n", blk.Number(), blk.Hash().String(), blk.ParentHash().String(), len(blk.Transactions), "BEGIN DISPUTE RESOLUTION")
	table.Reset()
}

func printSlashBlock(table *ansiterm.TabWriter, blk *edge_types.Block) {
	table.SetForeground(ansiterm.BrightRed)
	fmt.Fprintf(table, "%d\t%s\t%s\t%d\t%s\n", blk.Number(), blk.Hash().String(), blk.ParentHash().String(), len(blk.Transactions), "END DISPUTE RESOLUTION")
	table.Reset()
}

func isFraudProofBlock(blk *edge_types.Block) bool {
	_, exists := block.GetExtraDataFraudProofTarget(blk.Header)
	return exists
}

func isBeginDisputeResolutionBlock(blk *edge_types.Block) bool {
	_, exists := block.GetExtraDataBeginDisputeResolutionTarget(blk.Header)
	return exists
}

func isEndDisputeResolutionBlock(blk *edge_types.Block) bool {
	_, exists := block.GetExtraDataEndDisputeResolutionTarget(blk.Header)
	return exists
}
