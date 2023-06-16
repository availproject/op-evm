package availaccount

import (
	"log"
	"math/big"
	"math/rand"
	"os"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/spf13/cobra"

	"github.com/maticnetwork/avail-settlement/pkg/avail"
)

const (
	// AVL is a constant representing 1 Avail token in its smallest denomination.
	// The Avail token has 18 decimal places, so 1 Avail token equals 10^18 of its smallest denomination.
	AVL = 1_000_000_000_000_000_000

	// maxUint64 is the maximum value that can be represented by a uint64.
	maxUint64 = ^uint64(0)
)

// GetCommand returns a Cobra command for creating an avail account and depositing the balance.
// It takes no arguments and returns a pointer to a cobra.Command.
// Example usage:
// cmd := GetCommand()
//
//	if err := cmd.Execute(); err != nil {
//	   log.Fatalf("cmd.Execute error: %v", err)
//	}
func GetCommand() *cobra.Command {
	var balance uint64
	var availAddr, path string
	var retry bool
	cmd := &cobra.Command{
		Use:   "availaccount",
		Short: "Create an avail account and deposit the balance",
		Run: func(cmd *cobra.Command, args []string) {
			Run(availAddr, path, balance, retry)
		},
	}
	cmd.Flags().StringVar(&availAddr, "avail-addr", "ws://127.0.0.1:9944/v1/json-rpc", "Avail JSON-RPC URL")
	cmd.Flags().StringVar(&path, "path", "./configs/account", "Save path for account memonic file")
	cmd.Flags().Uint64Var(&balance, "balance", 18, "Number of AVLs to deposit on the account")
	cmd.Flags().BoolVar(&retry, "retry", false, "Retry if account deposit fails")
	return cmd
}

// Run is responsible for setting up and executing the process of creating an Avail account and
// depositing a balance into it. It takes the Avail JSON-RPC URL, a file path for saving the
// account mnemonic, the balance to deposit into the account, and a retry flag to indicate
// whether the process should be retried if an error occurs.
// Example usage:
// Run("ws://127.0.0.1:9944/v1/json-rpc", "./configs/account", 18, false)
func Run(availAddr, path string, balance uint64, retry bool) {
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

// deposit is a helper function used to deposit a specified balance into an Avail account.
// It works by depositing the balance in chunks of maxUint64 due to the API limitations.
// This function takes an Avail client, an Avail account, and a balance, and returns an error.
// Example usage (assuming availClient and availAccount are already defined):
//
//	if err := deposit(availClient, availAccount, 1000); err != nil {
//	   log.Fatalf("deposit error: %v", err)
//	}
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
