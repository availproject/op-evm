package tests

import (
	"context"
	"flag"
	"net/netip"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"

	"github.com/availproject/op-evm/consensus/avail"
	"github.com/availproject/op-evm/pkg/devnet"
)

// nolint:unused
var availAddr = flag.String("avail-addr", "ws://127.0.0.1:9944/v1/json-rpc", "Avail JSON-RPC URL")

// nolint:unused
var bindInterface = flag.String("bind-addr", "127.0.0.1", "IP address of the interface to bind node ports to")

// nolint:unused
var accountPath = flag.String("account-path", "../data/test-accounts", "Path to the account mnemonic file")

// nolint:unused
var bootnodeAddr = flag.String("bootnode-addr", "", "Remote bootstrap sequencer address")

func Test_MultipleSequencers(t *testing.T) {
	t.Skip("multi-sequencer e2e tests disabled in CI/CD due to lack of Avail")

	flag.Parse()

	var ethClient *ethclient.Client
	var err error

	if *bootnodeAddr == "" {
		t.Log("starting nodes")
		bindAddr, err := netip.ParseAddr(*bindInterface)
		if err != nil {
			t.Fatal(err)
		}

		ctx, err := devnet.StartNodes(hclog.Default(), bindAddr, *availAddr, *accountPath, avail.BootstrapSequencer, avail.Sequencer, avail.Sequencer, avail.WatchTower)
		if err != nil {
			t.Fatal(err)
		}

		// Shutdown all nodes once test finishes.
		t.Cleanup(ctx.StopAll)

		t.Log("nodes started")

		ethClient, err = ctx.GethClient(avail.BootstrapSequencer)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		ethClient, err = ethclient.Dial(*bootnodeAddr)
		if err != nil {
			t.Fatal(err)
		}
	}

	waitForPeers(t, ethClient, 3)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		bNum, err := ethClient.BlockNumber(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// Cleanup timeout context.
		cancel()

		// Wait for 5 blocks
		if bNum > 4 {
			break
		}

		time.Sleep(time.Second)
	}
}

// nolint:unused
func waitForPeers(t testing.TB, ethClient *ethclient.Client, minNodes int) {
	t.Helper()

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

		peerCount, err := ethClient.PeerCount(ctx)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("Got peer count: %d", peerCount)

		// Cleanup timeout context.
		cancel()

		if int(peerCount) >= minNodes {
			return
		}

		time.Sleep(250 * time.Millisecond)
	}
}
