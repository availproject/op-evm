package test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"reflect"
	"testing"

	edge_crypto "github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
	"github.com/maticnetwork/avail-settlement/pkg/test"
)

func Test_Builder_Construction_FromParentHash(t *testing.T) {
	executor, bchain := test.NewBlockchain(t, staking.NewVerifier(new(test.DumbActiveSequencers), hclog.Default()), getGenesisBasePath())
	h := bchain.Genesis()

	bbf := block.NewBlockBuilderFactory(bchain, executor, hclog.Default())
	_, err := bbf.FromParentHash(h)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Builder_Defaults(t *testing.T) {
	sk := newPrivateKey(t)

	_, err := newBlockBuilder(t).
		SignWith(sk).
		Build()

	if err != nil {
		t.Fatal(err)
	}
}

func Test_Builder_Set_Block_Number(t *testing.T) {
	sk := newPrivateKey(t)

	expected := uint64(42)

	b, err := newBlockBuilder(t).
		SetBlockNumber(expected).
		SignWith(sk).
		Build()

	if err != nil {
		t.Fatal(err)
	}

	if b.Header.Number != expected {
		t.Fatalf("block header number: got %d, expected %d", b.Header.Number, expected)
	}
}

func Test_Builder_Change_CoinbaseAddress(t *testing.T) {
	sk := newPrivateKey(t)

	var coinbase types.Address
	{
		ext := newPrivateKey(t)
		pk := ext.Public().(*ecdsa.PublicKey)
		coinbase = edge_crypto.PubKeyToAddress(pk)
	}

	b, err := newBlockBuilder(t).
		SetCoinbaseAddress(coinbase).
		SignWith(sk).
		Build()

	if err != nil {
		t.Fatal(err)
	}

	miner := types.BytesToAddress(b.Header.Miner)
	if miner != coinbase {
		t.Fatalf("block miner address: got %q, expected %q", miner, coinbase)
	}
}

func Test_Builder_Set_Difficulty(t *testing.T) {
	sk := newPrivateKey(t)

	expected := uint64(42)

	b, err := newBlockBuilder(t).
		SetDifficulty(expected).
		SignWith(sk).
		Build()

	if err != nil {
		t.Fatal(err)
	}

	if b.Header.Difficulty != expected {
		t.Fatalf("block header difficulty: got %d, expected %d", b.Header.Difficulty, expected)
	}
}

func Test_Builder_Change_GasLimit(t *testing.T) {
	sk := newPrivateKey(t)

	// Set ~correct gas limit.
	_, err := newBlockBuilder(t).
		SetGasLimit(4_713_000).
		SignWith(sk).
		Build()

	if err != nil {
		t.Fatal(err)
	}
}

func Test_Builder_Change_ParentStateRoot(t *testing.T) {
	sk := newPrivateKey(t)

	parentRoot := types.Hash{
		0x21, 0x21, 0x42, 0x42, 0xff, 0xff, 0xaf, 0xbe,
		0x83, 0xad, 0x12, 0xf1, 0xba, 0xb3, 0x02, 0x18,
		0xfe, 0xfe, 0x82, 0x83, 0xfe, 0xfe, 0x00, 0x00,
		0xaa, 0x1f, 0xc9, 0xc0, 0xe8, 0xc3, 0x00, 0x01,
	}

	_, err := newBlockBuilder(t).
		SetParentStateRoot(parentRoot).
		SignWith(sk).
		Build()

	if err == nil {
		t.Fatal(err)
	}
}

func Test_Builder_Add_Transaction(t *testing.T) {
	executor, bchain := test.NewBlockchain(t, staking.NewVerifier(new(test.DumbActiveSequencers), hclog.Default()), getGenesisBasePath())
	address, privateKey := test.NewAccount(t)
	address2, _ := test.NewAccount(t)

	// Deposit 100 ETH to first account.
	test.DepositBalance(t, address, big.NewInt(0).Mul(big.NewInt(100), test.ETH), bchain, executor)

	// Construct block.Builder w/ the blockchain instance that contains
	// balance for our test account.
	bbf := block.NewBlockBuilderFactory(bchain, executor, hclog.Default())
	bb, err := bbf.FromParentHash(bchain.Header().Hash)
	if err != nil {
		t.Fatal(err)
	}

	amount := big.NewInt(0).Mul(big.NewInt(10), test.ETH)

	// Transfer 10 ETH from first account to second one.
	tx := &types.Transaction{
		From:     address,
		To:       &address2,
		Value:    amount,
		Gas:      100000,
		GasPrice: big.NewInt(1),
	}

	err = bb.
		AddTransactions(tx).
		SignWith(privateKey).
		Write("test")

	if err != nil {
		t.Fatal(err)
	}

	blk, found := bchain.GetBlockByHash(bchain.Header().Hash, true)
	if !found {
		t.Fatalf("blockchain couldn't find header block")
	}

	numTxs := len(blk.Transactions)
	if numTxs != 1 {
		t.Fatalf("expected HEAD block to contain 1 transaction, found %d", numTxs)
	}

	tx = blk.Transactions[0]
	if tx.From != address {
		t.Fatalf("expected %q, got %q in tx.From", address, tx.From)
	}

	if *tx.To != address2 {
		t.Fatalf("expected %q, got %q in tx.To", address2, tx.To)
	}

	if tx.Value.Cmp(amount) != 0 {
		t.Fatalf("expected %q, got %q in tx.Value", tx.Value, amount)
	}
}

func Test_Builder_Set_ExtraData(t *testing.T) {
	sk := newPrivateKey(t)

	key := "foo"
	value := []byte("bar")

	b, err := newBlockBuilder(t).
		SetExtraDataField(key, value).
		SignWith(sk).
		Build()

	if err != nil {
		t.Fatal(err)
	}

	kv, err := block.DecodeExtraDataFields(b.Header.ExtraData)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(kv["foo"], value) {
		t.Fatalf("block header ExtraData[%q]: got %v, expected %v", "foo", kv["foo"], value)
	}
}

func newPrivateKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	return privateKey
}

func newBlockBuilder(t *testing.T) block.Builder {
	t.Helper()

	executor, bchain := test.NewBlockchain(t, staking.NewVerifier(new(test.DumbActiveSequencers), hclog.Default()), getGenesisBasePath())
	h := bchain.Genesis()

	bbf := block.NewBlockBuilderFactory(bchain, executor, hclog.Default())
	bb, err := bbf.FromParentHash(h)
	if err != nil {
		t.Fatal(err)
	}

	return bb
}
