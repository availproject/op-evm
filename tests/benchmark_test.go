package tests

import (
	"context"
	"flag"
	"math/big"
	"net/netip"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"

	"github.com/availproject/avail-settlement-contracts/testing/pkg/testtoken"
	"github.com/availproject/op-evm/consensus/avail"
	"github.com/availproject/op-evm/pkg/devnet"
)

const walletsDir = "../data/wallets"

func Benchmark_SendingTransactions(b *testing.B) {
	b.Skip("multi-sequencer benchmarks disabled in CI/CD due to lack of Avail")

	flag.Parse()

	ks := keystore.NewKeyStore(walletsDir, keystore.StandardScryptN, keystore.StandardScryptP)
	addr, err := netip.ParseAddr(*bindInterface)
	if err != nil {
		b.Fatal(err)
	}

	var ethClient *ethclient.Client
	if *bootnodeAddr == "" {
		b.Log("starting nodes")
		ctx, err := devnet.StartNodes(hclog.Default(), addr, *availAddr, *accountPath, avail.BootstrapSequencer, avail.Sequencer, avail.Sequencer, avail.WatchTower)
		if err != nil {
			b.Fatal(err)
		}
		// Shutdown all nodes once test finishes.
		b.Cleanup(ctx.StopAll)

		b.Log("nodes started")

		ethClient, err = ctx.GethClient(avail.BootstrapSequencer)
		if err != nil {
			b.Fatal(err)
		}
	} else {
		ethClient, err = ethclient.Dial(*bootnodeAddr)
		if err != nil {
			b.Fatal(err)
		}
	}

	waitForPeers(b, ethClient, 3)

	accs := ks.Accounts()
	ownerAccount := accs[0]
	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		b.Fatal(err)
	}
	auth, err := authOpts(ethClient, chainID, ks, ownerAccount)
	if err != nil {
		b.Fatal(err)
	}
	_, _, testToken, err := testtoken.DeployTesttoken(auth, ethClient)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("TestToken.mint", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err = testToken.Mint(auth, ownerAccount.Address, big.NewInt(1))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// nolint:unused
func authOpts(client *ethclient.Client, chainID *big.Int, ks *keystore.KeyStore, fromAccount accounts.Account) (*bind.TransactOpts, error) {
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	passpharse := "secret"
	err = ks.Unlock(fromAccount, passpharse)
	if err != nil {
		return nil, err
	}

	keyjson, err := ks.Export(fromAccount, passpharse, passpharse)
	if err != nil {
		return nil, err
	}

	privatekey, err := keystore.DecryptKey(keyjson, passpharse)
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privatekey.PrivateKey, chainID)
	if err != nil {
		return nil, err
	}
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(700000) // in units
	auth.GasPrice = gasPrice

	return auth, nil
}
