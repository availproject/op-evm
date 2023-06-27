package tests

import (
	"context"
	"flag"
	"github.com/availproject/op-evm/pkg/devnet"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
	"net/netip"
	"testing"
	"time"

	"github.com/availproject/op-evm/consensus/avail"
)

func Test_Fraud(t *testing.T) {
	t.Skip("fraud e2e tests disabled in CI/CD due to lack of Avail")

	flag.Parse()

	t.Log("starting nodes")

	bindAddr, err := netip.ParseAddr(*bindInterface)
	if err != nil {
		t.Fatal(err)
	}
	var ethClient *ethclient.Client
	if *bootnodeAddr == "" {
		ctx, err := devnet.StartNodes(hclog.Default(), bindAddr, *availAddr, *accountPath, avail.BootstrapSequencer, avail.Sequencer, avail.WatchTower)
		if err != nil {
			t.Fatal(err)
		}

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

	waitForPeers(t, ethClient, 2)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		bNum, err := ethClient.BlockNumber(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// Cleanup timeout context.
		cancel()

		// Wait for 10 blocks
		if bNum > 20 {
			t.Fatal("Could not receive successful confirmation that fraud block was processed in 10 blocks.")
		}

		time.Sleep(time.Second)
	}
}
