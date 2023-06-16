package tail

import (
	"fmt"

	edge_types "github.com/0xPolygon/polygon-edge/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/hashicorp/go-hclog"
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
	cmd.Flags().Int64Var(&offset, "offset", -1, "Block offset; defaults to first block")
	return cmd
}

func Run(availAddr string, offset int64) {
	availClient, err := avail.NewClient(availAddr)
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

	blk, err := availClient.SearchBlock(0, func(blk *types.SignedBlock) (int64, bool, error) {
		blks, err := block.FromAvail(blk, appID, callIdx, hclog.NewNullLogger())
		if err != nil {
			panic(err)
		}

		var min, max uint64
		for _, b := range blks {
			if b.Number() < min || min == 0 {
				min = b.Number()
			}

			if b.Number() > max || max == 0 {
				max = b.Number()
			}

			if b.Header.Number == uint64(offset) {
				return int64(blk.Block.Header.Number), true, nil
			}
		}

		// TODO: Fix the condition when the exact offset block is not found.
		return 0, false, fmt.Errorf("TODO")
	})

	availBlkStream := availClient.BlockStream(uint64(blk.Block.Header.Number))

	for blk := range availBlkStream.Chan() {
		blks, err := block.FromAvail(blk, appID, callIdx, hclog.NewNullLogger())
		if err != nil {
			panic(err)
		}

		for _, b := range blks {
			if b.Number() < uint64(offset) {
				continue
			}

			printBlock(b)
		}
	}
}

func printBlock(blk *edge_types.Block) {
	fmt.Printf("blk %d:%s\n", blk.Number(), blk.Hash().String())
}
