package main

import (
	"log"

	"github.com/0xPolygon/polygon-edge/command/secrets"
	"github.com/spf13/cobra"

	"github.com/availproject/op-evm/cmd/availaccount"
	"github.com/availproject/op-evm/cmd/devnet"
	"github.com/availproject/op-evm/cmd/server"
	"github.com/availproject/op-evm/cmd/tail"
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
