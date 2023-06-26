package tests

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
	"math/big"
	"math/rand"
	"net/netip"
	"testing"

	"github.com/availproject/op-evm-contracts/testing/pkg/testtoken"
	"github.com/availproject/op-evm/consensus/avail"
	"github.com/availproject/op-evm/pkg/devnet"
)

// privateKeyBytes faucet account private key
const privateKeyBytes = "e29fc399e151b829ca68ba811108965aeec52c21f2ac1744cb28f203231dc085"

func Benchmark_SendingTransactions(b *testing.B) {
	//b.Skip("multi-sequencer benchmarks disabled in CI/CD due to lack of Avail")

	flag.Parse()

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

	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		b.Fatal(err)
	}

	privatekey, err := crypto.BytesToECDSAPrivateKey([]byte(privateKeyBytes))
	if err != nil {
		b.Fatal(err)
	}
	address := common.Address(crypto.PubKeyToAddress(privatekey.Public().(*ecdsa.PublicKey)))

	nonce, err := ethClient.PendingNonceAt(context.Background(), address)
	if err != nil {
		b.Fatal(err)
	}

	auth, err := authOpts(ethClient, chainID, privatekey)
	if err != nil {
		b.Fatal(err)
	}
	auth.Nonce = big.NewInt(0).SetUint64(nonce + 1)
	_, _, testToken, err := testtoken.DeployTesttoken(auth, ethClient)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("TestToken.mint", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err = testToken.Mint(auth, address, big.NewInt(rand.Int63n(10000000)+1))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// nolint:unused
func authOpts(client *ethclient.Client, chainID *big.Int, privatekey *ecdsa.PrivateKey) (*bind.TransactOpts, error) {
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privatekey, chainID)
	if err != nil {
		return nil, err
	}
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(700000) // in units
	auth.GasPrice = gasPrice

	return auth, nil
}
