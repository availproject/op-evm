package tail

import (
	"errors"
	"fmt"
	"os"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	"github.com/juju/ansiterm"
	"github.com/spf13/cobra"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"

	"github.com/availproject/op-evm/pkg/avail"
	"github.com/availproject/op-evm/pkg/block"
	"github.com/availproject/op-evm/pkg/staking"
)

func GetCommand() *cobra.Command {
	var availAddr, jsonrpcAddr string
	var offset int64
	cmd := &cobra.Command{
		Use:   "tail",
		Short: "Follow Avail SL blockstream from Avail",
		Run: func(cmd *cobra.Command, args []string) {
			Run(availAddr, jsonrpcAddr, offset)
		},
	}
	cmd.Flags().StringVar(&availAddr, "avail-addr", "ws://127.0.0.1:9944/v1/json-rpc", "Avail JSON-RPC URL")
	cmd.Flags().StringVar(&jsonrpcAddr, "jsonrpc-addr", "http://127.0.0.1:10002/v1/json-rpc", "Optimistic EVM Rollup JSON-RPC URL")
	cmd.Flags().Int64Var(&offset, "offset", 1, "Block offset; defaults to first block")
	return cmd
}

func Run(availAddr, jsonrpcAddr string, offset int64) {
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

	jsonrpcClnt, err := jsonrpc.NewClient(jsonrpcAddr)
	if err != nil {
		panic(err)
	}

	availBlkStream := availClient.BlockStream(1)

	tw := ansiterm.NewTabWriter(os.Stdout, 4, 4, 1, ' ', 0)

	for blk := range availBlkStream.Chan() {
		blks, err := avail.BlockFromAvail(blk, appID, callIdx, hclog.NewNullLogger())
		if err != nil && !errors.Is(err, avail.ErrNoExtrinsicFound) {
			panic(err)
		}

		for _, b := range blks {
			if b.Number() < uint64(offset) {
				continue
			}

			printBlock(tw, jsonrpcClnt, b)
		}

		tw.Flush()
	}
}

func printBlock(tw *ansiterm.TabWriter, jsonrpcClnt *jsonrpc.Client, blk *types.Block) {
	switch {
	case blk.Number() == 0:
		printGenesis(tw, blk)
	case isFraudProofBlock(blk):
		printFraudProofBlock(tw, blk)
	case isBeginDisputeResolutionBlock(blk):
		printBeginDisputeResolutionBlock(tw, jsonrpcClnt, blk)
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
	targetBlockHash, _ := block.GetExtraDataFraudProofTarget(blk.Header)
	tw.SetForeground(ansiterm.BrightYellow)
	fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%s\n", blk.Number(), blk.Hash().String(), blk.ParentHash().String(), len(blk.Transactions), fmt.Sprintf("FRAUDPROOF (TARGET BLOCK: %s)", targetBlockHash.String()))
	tw.Reset()
}

func printBeginDisputeResolutionBlock(tw *ansiterm.TabWriter, jsonrpcClnt *jsonrpc.Client, blk *types.Block) {
	parent, err := jsonrpcClnt.Eth().GetBlockByHash(ethgo.Hash(blk.ParentHash()), true)
	if err != nil {
		panic(err)
	}

	isFork := parent.Number+1 != blk.Number()

	if isFork {
		tw.SetForeground(ansiterm.BrightGreen)
		fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%s\n", blk.Number(), blk.Hash().String(), blk.ParentHash().String(), len(blk.Transactions), "BEGIN DISPUTE RESOLUTION")
		tw.Flush()
		fmt.Fprintf(tw, "\t%s\t%d -> %d\t%s -> %s\n", "↳ FORK:", parent.Number, blk.Number(), blk.ParentHash().String(), blk.Hash().String())
	} else {
		tw.SetForeground(ansiterm.BrightBlue)
		fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%s\n", blk.Number(), blk.Hash().String(), blk.ParentHash().String(), len(blk.Transactions), "BEGIN DISPUTE RESOLUTION")
	}

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
	var isBeginDisputeResolutionBlock bool
	for _, tx := range blk.Transactions {
		if !isBeginDisputeResolutionBlock {
			// Ignore returned error on purpose. It's mostly for debugging purposes and useless here.
			isBeginDisputeResolutionBlock, _ = staking.IsBeginDisputeResolutionTx(tx)
		} else {
			break
		}
	}
	return isBeginDisputeResolutionBlock
}

func isEndDisputeResolutionBlock(blk *types.Block) bool {
	_, exists := block.GetExtraDataEndDisputeResolutionTarget(blk.Header)
	return exists
}
