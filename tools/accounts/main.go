package main

import (
	"flag"
	"log"
	"os"

	"github.com/maticnetwork/avail-settlement/pkg/avail"
)

const (
	// 1 AVL == 10^18 Avail fractions.
	AVL = 1_000_000_000_000_000_000
)

func main() {

	var balance uint64
	var availAddr, path string
	flag.StringVar(&availAddr, "avail-addr", "ws://127.0.0.1:9944/v1/json-rpc", "Avail JSON-RPC URL")
	flag.StringVar(&path, "path", "./configs/account", "Save path for account memonic file")
	flag.Uint64Var(&balance, "balance", 18, "Path to the configuration file")

	flag.Parse()

	availClient, err := avail.NewClient(availAddr)
	if err != nil {
		panic(err)
	}

	log.Print("Creating new avail account...")

	availAccount, err := avail.NewAccount()
	if err != nil {
		panic(err)
	}

	log.Printf("Created new avail account %+v", availAccount)
	log.Printf("Depositing %d AVL to '%s'...", balance, availAccount.Address)

	err = avail.DepositBalance(availClient, availAccount, balance*AVL, 0)
	if err != nil {
		panic(err)
	}

	log.Printf("Successfully deposited '%d' AVL to '%s'", balance, availAccount.Address)

	if err := os.WriteFile(path, []byte(availAccount.URI), 0644); err != nil {
		panic(err)
	}

	log.Printf("Successfuly written mnemonic into '%s'", path)
}
