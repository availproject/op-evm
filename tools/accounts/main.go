package main

import (
	"flag"
	"log"
	"math/big"
	"math/rand"
	"os"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"

	"github.com/maticnetwork/avail-settlement/pkg/avail"
)

const (
	// 1 AVL == 10^18 Avail fractions.
	AVL = 1_000_000_000_000_000_000

	maxUint64 = ^uint64(0)
)

func main() {
	var balance uint64
	var availAddr, path string
	var retry bool
	flag.StringVar(&availAddr, "avail-addr", "ws://127.0.0.1:9944/v1/json-rpc", "Avail JSON-RPC URL")
	flag.StringVar(&path, "path", "./configs/account", "Save path for account memonic file")
	flag.Uint64Var(&balance, "balance", 18, "Number of AVLs to deposit on the account")
	flag.BoolVar(&retry, "retry", false, "Retry if account deposit fails")

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

	if retry {
		for {
			if err = deposit(availClient, availAccount, balance); err == nil {
				break
			}

			seconds := 20 + rand.Intn(11)
			log.Println("ERROR: ", err)
			log.Printf("Creating avail account and deposit tokens failed, retrying in %d seconds...\n", seconds)
			time.Sleep(time.Duration(seconds) * time.Second)
		}
	} else {
		if err = deposit(availClient, availAccount, balance); err != nil {
			panic(err)
		}
	}

	log.Printf("Successfully deposited '%d' AVL to '%s'", balance, availAccount.Address)

	if err := os.WriteFile(path, []byte(availAccount.URI), 0o644); err != nil {
		panic(err)
	}

	log.Printf("Successfuly written mnemonic into '%s'", path)
}

// Deposit balance in chunks of maxUint64 (because of the API limitations).
func deposit(availClient avail.Client, availAccount signature.KeyringPair, balance uint64) (err error) {
	amount := big.NewInt(0).Mul(big.NewInt(0).SetUint64(balance), big.NewInt(AVL))

	for {
		if amount.IsUint64() {
			err = avail.DepositBalance(availClient, availAccount, amount.Uint64(), 0)
			if err != nil {
				return err
			}

			break
		} else {
			err = avail.DepositBalance(availClient, availAccount, maxUint64, 0)
			if err != nil {
				return err
			}

			amount = big.NewInt(0).Sub(amount, big.NewInt(0).SetUint64(maxUint64))
		}
	}
	return
}
