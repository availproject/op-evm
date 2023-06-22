package main

import (
	"log"

	"github.com/0xPolygon/polygon-edge/command/secrets"
	"github.com/spf13/cobra"

	"github.com/maticnetwork/avail-settlement/cmd/availaccount"
	"github.com/maticnetwork/avail-settlement/cmd/devnet"
	"github.com/maticnetwork/avail-settlement/cmd/server"
	"github.com/maticnetwork/avail-settlement/cmd/tail"
)

func main() {
	cmd := &cobra.Command{
		Short: "Avail settlement layer",
	}
	cmd.AddCommand(
		server.GetCommand(),
		availaccount.GetCommand(),
		devnet.GetCommand(),
		secrets.GetCommand(),
		tail.GetCommand(),
	)
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
