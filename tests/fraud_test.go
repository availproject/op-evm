package tests

import (
	"context"
	"flag"
	"net/netip"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/maticnetwork/avail-settlement/consensus/avail"
)

func Test_Fraud(t *testing.T) {
	//t.Skip("fraud e2e tests disabled in CI/CD due to lack of Avail")

	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	*genesisCfgPath = filepath.Join(cwd, *genesisCfgPath)

	t.Log("starting nodes")

	bindAddr, err := netip.ParseAddr(*bindInterface)
	if err != nil {
		t.Fatal(err)
	}

	ctx, err := StartNodes(t, bindAddr, *genesisCfgPath, *availAddr, *accountPath, avail.BootstrapSequencer, avail.Sequencer, avail.WatchTower)
	if err != nil {
		t.Fatal(err)
	}

	// Shutdown all nodes once test finishes.
	t.Cleanup(ctx.StopAll)

	t.Log("nodes started")

	ethClient, err := ctx.GethClient()
	if err != nil {
		t.Fatal(err)
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
