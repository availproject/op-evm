package server

import (
	"errors"
	"log"
	"time"

	"github.com/0xPolygon/polygon-edge/helper/common"
	"github.com/hashicorp/go-hclog"
	golog "github.com/ipfs/go-log/v2"
	"github.com/spf13/cobra"

	consensus "github.com/availproject/op-evm/consensus/avail"
	"github.com/availproject/op-evm/pkg/avail"
	"github.com/availproject/op-evm/pkg/config"
	"github.com/availproject/op-evm/server"
)

// GetCommand returns a Cobra command for running the settlement layer server.
// It takes no arguments and returns a pointer to a cobra.Command.
// Example usage:
// cmd := GetCommand()
//
//	if err := cmd.Execute(); err != nil {
//	   log.Fatalf("cmd.Execute error: %v", err)
//	}
func GetCommand() *cobra.Command {
	var bootnode bool
	var availAddr, path, accountPath, fraudListenAddr string
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Run the settlement layer server",
		Run: func(cmd *cobra.Command, args []string) {
			Run(availAddr, path, accountPath, fraudListenAddr, bootnode)
		},
	}
	cmd.Flags().StringVar(&availAddr, "avail-addr", "ws://127.0.0.1:9944/v1/json-rpc", "Avail JSON-RPC URL")
	cmd.Flags().StringVar(&path, "config-file", "./configs/bootnode.yaml", "Path to the configuration file")
	cmd.Flags().StringVar(&accountPath, "account-config-file", "./configs/account", "Path to the account mnemonic file")
	cmd.Flags().BoolVar(&bootnode, "bootstrap", false, "bootstrap flag must be specified for the first node booting a new network from the genesis")
	cmd.Flags().StringVar(&fraudListenAddr, "fraud-srv-listen-addr", ":9990", "Fraud server listen address")
	return cmd
}

// Run initializes and starts the settlement layer server. It takes the Avail JSON-RPC URL, a file path for
// the configuration file, a file path for the account mnemonic file, a fraud server listen address and a bootnode
// flag. It does not return a value.
// Example usage:
// Run("ws://127.0.0.1:9944/v1/json-rpc", "./configs/bootnode.yaml", "./configs/account", ":9990", false)
func Run(availAddr, path, accountPath, fraudListenAddr string, bootnode bool) {
	// Enable LibP2P logging but only >= warn
	golog.SetAllLoggers(golog.LevelWarn)

	config, err := config.NewServerConfig(path)
	if err != nil {
		log.Fatalf("failure to get node configuration: %s", err)
	}

	// Enable TxPool P2P gossiping
	config.Config.Seal = true

	availAccount, err := avail.AccountFromFile(accountPath)
	if err != nil {
		log.Fatalf("failed to read Avail account from %q: %s\n", accountPath, err)
	}

	availClient, err := avail.NewClient(availAddr, hclog.Default())
	if err != nil {
		log.Fatalf("failed to create Avail client: %s\n", err)
	}

	appID, err := avail.EnsureApplicationKeyExists(availClient, avail.ApplicationKey, availAccount)
	if err != nil {
		log.Fatalf("failed to get AppID from Avail: %s\n", err)
	}

	availSender := avail.NewSender(availClient, appID, availAccount)

	cfg := consensus.Config{
		AvailAccount:      availAccount,
		AvailClient:       availClient,
		AvailSender:       availSender,
		Bootnode:          bootnode,
		FraudListenerAddr: fraudListenAddr,
		NodeType:          config.NodeType,
		AvailAppID:        appID,
	}
	serverInstance, err := server.NewServer(config.Config, cfg)
	if err != nil {
		log.Fatalf("failure to start node: %s", err)
	}

	if err := HandleSignals(serverInstance.Close); err != nil {
		log.Fatalf("handle signal error: %s", err)
	}
}

// HandleSignals is a function that handles signals sent to the console. It helps in managing
// the lifecycle of the server by triggering a shutdown when a termination signal is received.
// It takes a function to be called when a termination signal is received and returns an error if
// the server shutdown was not graceful.
// Example usage (assuming serverInstance is already defined):
//
//	if err := HandleSignals(serverInstance.Close); err != nil {
//	   log.Fatalf("handle signal error: %v", err)
//	}
func HandleSignals(closeFn func()) error {
	signalCh := common.GetTerminationSignalCh()
	sig := <-signalCh

	log.Printf("\n[SIGNAL] Caught signal: %v\n", sig)

	// Call the Minimal server close callback
	gracefulCh := make(chan struct{})

	go func() {
		if closeFn != nil {
			closeFn()
		}

		close(gracefulCh)
	}()

	select {
	case <-signalCh:
		return errors.New("shutdown by signal channel")
	case <-time.After(5 * time.Second):
		return errors.New("shutdown by timeout")
	case <-gracefulCh:
		return nil
	}
}
