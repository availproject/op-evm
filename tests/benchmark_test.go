package tests

import (
	"context"
	"flag"
	"log"
	"math/big"
	"net/netip"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/maticnetwork/avail-settlement-contracts/testing/pkg/testtoken"
	"github.com/maticnetwork/avail-settlement/consensus/avail"
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

	b.Log("starting nodes")
	ctx, err := StartNodes(b, addr, *genesisCfgPath, *availAddr, avail.BootstrapSequencer, avail.Sequencer, avail.Sequencer, avail.WatchTower)
	if err != nil {
		b.Fatal(err)
	}

	// Shutdown all nodes once test finishes.
	b.Cleanup(ctx.StopAll)

	b.Log("nodes started")

	ethClient, err := ctx.GethClient()
	if err != nil {
		b.Fatal(err)
	}

	waitForPeers(b, ethClient, 3)

	accs := ks.Accounts()
	ownerAccount := accs[0]
	url, err := ctx.FirstRPCURLForNodeType(avail.Sequencer)
	if err != nil {
		log.Fatal(err)
	}
	sequencerClient, err := ethclient.Dial(url.String())
	if err != nil {
		b.Fatal(err)
	}
	chainID, err := sequencerClient.ChainID(context.Background())
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
			auth, err := authOpts(ethClient, chainID, ks, ownerAccount)
			if err != nil {
				b.Fatal(err)
			}
			_, err = testToken.Mint(auth, ownerAccount.Address, big.NewInt(1))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func authOpts(client *ethclient.Client, chainID *big.Int, ks *keystore.KeyStore, fromAccount accounts.Account) (*bind.TransactOpts, error) {
	nonce, err := client.PendingNonceAt(context.Background(), fromAccount.Address)
	if err != nil {
		return nil, err
	}

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
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(700000) // in units
	auth.GasPrice = gasPrice

	return auth, nil
}
