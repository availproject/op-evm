package server

import (
	"errors"
	"log"
	"time"

	"github.com/0xPolygon/polygon-edge/helper/common"
	golog "github.com/ipfs/go-log/v2"
	"github.com/spf13/cobra"

	consensus "github.com/maticnetwork/avail-settlement/consensus/avail"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
	"github.com/maticnetwork/avail-settlement/pkg/config"
	"github.com/maticnetwork/avail-settlement/server"
)

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

func Run(availAddr, path, accountPath, fraudListenAddr string, bootnode bool) {
	// Enable LibP2P logging
	golog.SetAllLoggers(golog.LevelWarn)

	config, err := config.NewServerConfig(path)
	if err != nil {
		log.Fatalf("failure to get node configuration: %s", err)
	}

	// Enable TxPool P2P gossiping
	config.Config.Seal = true

	log.Printf("Server config: %+v", config)

	availAccount, err := avail.AccountFromFile(accountPath)
	if err != nil {
		log.Fatalf("failed to read Avail account from %q: %s\n", accountPath, err)
	}

	availClient, err := avail.NewClient(availAddr)
	if err != nil {
		log.Fatalf("failed to create Avail client: %s\n", err)
	}

	appID, err := avail.EnsureApplicationKeyExists(availClient, avail.ApplicationKey, availAccount)
	if err != nil {
		log.Fatalf("failed to get AppID from Avail: %s\n", err)
	}

	availSender := avail.NewSender(availClient, appID, availAccount)

	// Attach the consensus to the server
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

	log.Printf("Server instance %#v", serverInstance)

	if err := HandleSignals(serverInstance.Close); err != nil {
		log.Fatalf("handle signal error: %s", err)
	}
}

// HandleSignals is a helper method for handling signals sent to the console
// Like stop, error, etc.
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
