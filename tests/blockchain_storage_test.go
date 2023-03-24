package tests

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/blockchain/storage"
	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/state"
	itrie "github.com/0xPolygon/polygon-edge/state/immutable-trie"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/common"
	"github.com/maticnetwork/avail-settlement/pkg/snapshot"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
	"github.com/maticnetwork/avail-settlement/pkg/test"
)

func newKey(t *testing.T) (*ecdsa.PrivateKey, types.Address) {
	key, _, err := crypto.GenerateAndEncodeECDSAPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	addr := crypto.PubKeyToAddress(&key.PublicKey)

	return key, addr
}

func Test_OptimisticStateTransfer(t *testing.T) {
	minerKey, minerAddress := newKey(t)
	key1, address1 := newKey(t)
	key2, address2 := newKey(t)

	instance := NewTestInstance(t, minerAddress, address1, address2)
	instance2 := NewTestInstance(t, minerAddress, address1, address2)

	minerBalance := instance.GetBalance(minerAddress)
	address1Balance := instance.GetBalance(address1)
	address2Balance := instance.GetBalance(address2)
	t.Logf("instance1[miner: %d, address1: %d, address2: %d]\n", minerBalance, address1Balance, address2Balance)

	minerBalance = instance2.GetBalance(minerAddress)
	address1Balance = instance2.GetBalance(address1)
	address2Balance = instance2.GetBalance(address2)
	t.Logf("instance2[miner: %d, address1: %d, address2: %d]\n", minerBalance, address1Balance, address2Balance)

	hdr1 := instance.GetHeader()
	hdr2 := instance2.GetHeader()
	diff := cmp.Diff(hdr1, hdr2)
	if diff != "" {
		t.Fatalf("first checkpoint: \n%s", diff)
	}

	// Node 1 block -> Node 2
	instance.BeginSnapshot()
	tx1 := transferTokensTx(t, 10, key1, address2)
	tx2 := transferTokensTx(t, 3, key2, address1)
	instance.ExecuteTransactions(minerKey, tx1, tx2)
	state, bchain := instance.EndSnapshot()

	instance2.ApplySnapshots(state, bchain)

	hdr1 = instance.GetHeader()
	hdr2 = instance2.GetHeader()
	diff = cmp.Diff(hdr1, hdr2)
	if diff != "" {
		t.Fatalf("second checkpoint: \n%s", diff)
	}

	minerBalance = instance.GetBalance(minerAddress)
	address1Balance = instance.GetBalance(address1)
	address2Balance = instance.GetBalance(address2)
	t.Logf("instance1[miner: %d, address1: %d, address2: %d]\n", minerBalance, address1Balance, address2Balance)

	minerBalance = instance2.GetBalance(minerAddress)
	address1Balance = instance2.GetBalance(address1)
	address2Balance = instance2.GetBalance(address2)
	t.Logf("instance2[miner: %d, address1: %d, address2: %d]\n", minerBalance, address1Balance, address2Balance)

	// Node 2 block -> Node 1
	instance2.BeginSnapshot()
	tx1 = transferTokensTx(t, 4, key2, address1)
	tx2 = transferTokensTx(t, 7, key1, address2)
	instance2.ExecuteTransactions(minerKey, tx1, tx2)
	state, bchain = instance2.EndSnapshot()
	instance.ApplySnapshots(state, bchain)

	hdr1 = instance.GetHeader()
	hdr2 = instance2.GetHeader()
	diff = cmp.Diff(hdr1, hdr2)
	if diff != "" {
		t.Fatalf("third checkpoint: \n%s", diff)
	}

	// Node 1 block -> Node 2
	instance.BeginSnapshot()
	tx1 = transferTokensTx(t, 10, key1, address2)
	tx2 = transferTokensTx(t, 3, key2, address1)
	instance.ExecuteTransactions(minerKey, tx1, tx2)
	state, bchain = instance.EndSnapshot()
	instance2.ApplySnapshots(state, bchain)

	hdr1 = instance.GetHeader()
	hdr2 = instance2.GetHeader()
	diff = cmp.Diff(hdr1, hdr2)
	if diff != "" {
		t.Fatalf("fourth checkpoint: \n%s", diff)
	}

	// Node 2 block -> Node 1
	instance2.BeginSnapshot()
	tx1 = transferTokensTx(t, 4, key2, address1)
	tx2 = transferTokensTx(t, 7, key1, address2)
	instance2.ExecuteTransactions(minerKey, tx1, tx2)
	state, bchain = instance2.EndSnapshot()
	instance.ApplySnapshots(state, bchain)

	hdr1 = instance.GetHeader()
	hdr2 = instance2.GetHeader()
	diff = cmp.Diff(hdr1, hdr2)
	if diff != "" {
		t.Fatalf("fifth checkpoint: \n%s", diff)
	}

	minerBalance = instance.GetBalance(minerAddress)
	address1Balance = instance.GetBalance(address1)
	address2Balance = instance.GetBalance(address2)
	t.Logf("instance1[miner: %d, address1: %d, address2: %d]\n", minerBalance, address1Balance, address2Balance)

	minerBalance = instance2.GetBalance(minerAddress)
	address1Balance = instance2.GetBalance(address1)
	address2Balance = instance2.GetBalance(address2)
	t.Logf("instance2[miner: %d, address1: %d, address2: %d]\n", minerBalance, address1Balance, address2Balance)

	// Execute same transactions on both
	tx1 = transferTokensTx(t, 1, minerKey, address1)
	tx2 = transferTokensTx(t, 1, minerKey, address2)
	instance.ExecuteTransactions(minerKey, tx1, tx2)
	instance2.ExecuteTransactions(minerKey, tx1, tx2)

	hdr1 = instance.GetHeader()
	hdr2 = instance2.GetHeader()
	diff = cmp.Diff(hdr1, hdr2)
	if diff != "" {
		t.Fatalf("sixth checkpoint: \n%s", diff)
	}
}

func transferTokensTx(t *testing.T, n int64, from *ecdsa.PrivateKey, to types.Address) *types.Transaction {
	t.Helper()

	fromAddr := crypto.PubKeyToAddress(&from.PublicKey)

	tx := &types.Transaction{
		From:     fromAddr,
		To:       &to,
		Value:    big.NewInt(n),
		GasPrice: big.NewInt(5000),
		Gas:      1_000_000,
	}

	txSigner := &crypto.FrontierSigner{}
	tx, err := txSigner.SignTx(tx, from)
	if err != nil {
		t.Fatal(err)
	}

	return tx
}

type testInstance struct {
	t *testing.T

	blockchain *blockchain.Blockchain
	executor   *state.Executor

	blockBuilderFactory   block.BlockBuilderFactory
	blockchainSnapshotter snapshot.BlockchainKVSnapshotter
	stateSnapshotter      snapshot.StateStorageSnapshotter
}

func NewTestInstance(t *testing.T, addrs ...types.Address) *testInstance {
	logger := hclog.Default()
	chainSpec := test.NewChain(t, getGenesisBasePath())

	for _, addr := range addrs {
		chainSpec.Genesis.Alloc[addr] = &chain.GenesisAccount{
			Balance: big.NewInt(0).Mul(big.NewInt(1000), common.ETH),
		}
	}

	stateSnapshotter := snapshot.StateWrapper(itrie.NewMemoryStorage())
	executor := state.NewExecutor(chainSpec.Params, itrie.NewState(stateSnapshotter), logger)

	gr := executor.WriteGenesis(chainSpec.Genesis.Alloc)
	chainSpec.Genesis.StateRoot = gr

	// use the eip155 signer
	signer := crypto.NewEIP155Signer(uint64(chainSpec.Params.ChainID))

	blockchainSnapshotter, err := snapshot.NewLevelDBSnapshotStorage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	bchainStorage := storage.NewKeyValueStorage(logger.Named("leveldb"), blockchainSnapshotter)
	bchain, err := blockchain.NewBlockchain(logger, bchainStorage, chainSpec, nil, executor, signer)
	if err != nil {
		t.Fatal(err)
	}

	bchain.SetConsensus(staking.NewVerifier(new(staking.DumbActiveParticipants), logger))
	executor.GetHash = bchain.GetHashHelper

	if err := bchain.ComputeGenesis(); err != nil {
		t.Fatal(err)
	}

	return &testInstance{
		t: t,

		blockchain: bchain,
		executor:   executor,

		blockBuilderFactory:   block.NewBlockBuilderFactory(bchain, executor, logger),
		blockchainSnapshotter: blockchainSnapshotter,
		stateSnapshotter:      stateSnapshotter,
	}
}

func (ti *testInstance) GetBlockBuilderFromHead() block.Builder {
	bldr, err := ti.blockBuilderFactory.FromBlockchainHead()
	if err != nil {
		ti.t.Fatal(err)
	}

	return bldr
}

func (ti *testInstance) BeginSnapshot() {
	ti.stateSnapshotter.Begin()
	ti.blockchainSnapshotter.Begin()
}

func (ti *testInstance) EndSnapshot() (snapshot.StateStorageSnapshot, snapshot.BlockchainSnapshot) {
	ss := ti.stateSnapshotter.End()
	bs := ti.blockchainSnapshotter.End()
	return ss, bs
}

func (ti *testInstance) ApplySnapshots(ss snapshot.StateStorageSnapshot, bs snapshot.BlockchainSnapshot) {
	err := ti.stateSnapshotter.Apply(ss)
	if err != nil {
		ti.t.Fatal(err)
	}

	ti.executor.State().ResetCache()

	err = ti.blockchainSnapshotter.Apply(bs)
	if err != nil {
		ti.t.Fatal(err)
	}

	err = ti.blockchain.ComputeGenesis()
	if err != nil {
		ti.t.Fatal(err)
	}
}

func (ti *testInstance) ExecuteTransactions(signKey *ecdsa.PrivateKey, txs ...*types.Transaction) {
	bb, err := ti.blockBuilderFactory.FromBlockchainHead()
	if err != nil {
		ti.t.Fatal(err)
	}

	coinbaseAddr := crypto.PubKeyToAddress(&signKey.PublicKey)

	bb.SetCoinbaseAddress(coinbaseAddr)
	bb.AddTransactions(txs...)
	bb.SignWith(signKey)

	err = bb.Write("test")
	if err != nil {
		ti.t.Fatal(err)
	}
}

func (ti *testInstance) GetHeader() *types.Header {
	return ti.blockchain.Header()
}

func (ti *testInstance) GetBalance(addr types.Address) int64 {
	hdr := ti.blockchain.Header()
	transition, err := ti.executor.BeginTxn(hdr.StateRoot, hdr, types.BytesToAddress(hdr.Miner))
	if err != nil {
		ti.t.Fatal(err)
	}

	b := transition.GetBalance(addr)
	return b.Int64()
}
