package devnet

import (
	_ "embed"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"net/netip"

	"github.com/maticnetwork/avail-settlement/cmd/server"
	"github.com/maticnetwork/avail-settlement/consensus/avail"
	"github.com/maticnetwork/avail-settlement/pkg/devnet"
)

func GetCommand() *cobra.Command {
	var nodesCount, watchtowerCount int
	var availAddr, bindInterface, accountsPath string
	cmd := &cobra.Command{
		Use:   "devnet",
		Short: "Run a devnet environment",
		Run: func(cmd *cobra.Command, args []string) {
			log := hclog.Default()
			ctx, err := Run(log, nodesCount, watchtowerCount, availAddr, bindInterface, accountsPath)
			if err != nil {
				log.Error("starting devnet error", "err", err)
				return
			}

			ctx.Output(cmd.OutOrStdout())
			if err := server.HandleSignals(ctx.StopAll); err != nil {
				log.Error("handle signal error: %w", "err", err)
				return
			}
		},
	}
	cmd.Flags().IntVarP(&nodesCount, "node-count", "n", 1, "The number of sequencers")
	cmd.Flags().IntVarP(&watchtowerCount, "watchtower-count", "w", 1, "The number of watchtowers")
	cmd.Flags().StringVar(&availAddr, "avail-addr", "ws://127.0.0.1:9944/v1/json-rpc", "Avail JSON-RPC URL")
	cmd.Flags().StringVar(&bindInterface, "bind-addr", "127.0.0.1", "IP address of the interface to bind node ports to")
	cmd.Flags().StringVar(&accountsPath, "account-path", "./data/test-accounts", "Path to the account mnemonic file")
	return cmd
}

func Run(log hclog.Logger, nodesCount, watchtowerCount int, availAddr, bindInterface, accountsPath string) (*devnet.Context, error) {
	log.Info("starting nodes")
	bindAddr, err := netip.ParseAddr(bindInterface)
	if err != nil {
		return nil, fmt.Errorf("unable to parse bind interface: %w", err)
	}
	nodeTypes := []avail.MechanismType{avail.BootstrapSequencer}
	for i := 0; i < nodesCount; i++ {
		nodeTypes = append(nodeTypes, avail.Sequencer)
	}
	for i := 0; i < watchtowerCount; i++ {
		nodeTypes = append(nodeTypes, avail.WatchTower)
	}
	ctx, err := devnet.StartNodes(log, bindAddr, availAddr, accountsPath, nodeTypes...)
	if err != nil {
		return nil, fmt.Errorf("unable to start devnet: %w", err)
	}
	return ctx, nil
}
