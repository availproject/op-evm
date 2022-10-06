package block

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"testing"

	edge_crypto "github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/test"
)

func Test_Builder_Construction_FromParentHash(t *testing.T) {
	executor, bchain := test.NewBlockchain(t, NewVerifier(hclog.Default()))
	h := bchain.Genesis()

	bbf := NewBlockBuilderFactory(bchain, executor, hclog.Default())
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

func Test_Builder_Change_Invalid_GasLimit(t *testing.T) {
	sk := newPrivateKey(t)

	// Set invalid gas limit.
	b, err := newBlockBuilder(t).
		SetGasLimit(1).
		SignWith(sk).
		Build()

	if b != nil && err == nil {
		t.Fatal("no error from block building despite of invalid gas limit")
	}
}

func Test_Builder_Change_Valid_GasLimit(t *testing.T) {
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
	sk := newPrivateKey(t)
	addr := addressFromPrivateKey(t, sk)

	tx := &types.Transaction{
		From:     addr,
		To:       &addr,
		Value:    big.NewInt(0).Mul(big.NewInt(10), test.ETH), // 10 ETH
		GasPrice: big.NewInt(10),
	}

	_, err := newBlockBuilder(t).
		AddTransactions(tx).
		SignWith(sk).
		Build()

	if err != nil {
		t.Fatal(err)
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

func addressFromPrivateKey(t *testing.T, privateKey *ecdsa.PrivateKey) types.Address {
	pk := privateKey.Public().(*ecdsa.PublicKey)
	address := edge_crypto.PubKeyToAddress(pk)
	return address
}

func newBlockBuilder(t *testing.T) Builder {
	t.Helper()

	executor, bchain := test.NewBlockchain(t, NewVerifier(hclog.Default()))
	h := bchain.Genesis()

	bbf := NewBlockBuilderFactory(bchain, executor, hclog.Default())
	bb, err := bbf.FromParentHash(h)
	if err != nil {
		t.Fatal(err)
	}

	return bb
}
