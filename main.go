package main

import (
	"log"

	"github.com/0xPolygon/polygon-edge/command/secrets"
	"github.com/spf13/cobra"

	"github.com/maticnetwork/avail-settlement/cmd/availaccount"
	"github.com/maticnetwork/avail-settlement/cmd/devnet"
	"github.com/maticnetwork/avail-settlement/cmd/server"
)

func main() {
	var cmd = &cobra.Command{
		Short: "Avail settlement layer",
	}
	cmd.AddCommand(
		server.GetCommand(),
		availaccount.GetCommand(),
		devnet.GetCommand(),
		secrets.GetCommand(),
	)
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
