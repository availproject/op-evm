package tests

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	edge_crypto "github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/availproject/op-evm/pkg/block"
	"github.com/availproject/op-evm/pkg/common"
	"github.com/availproject/op-evm/pkg/staking"
	"github.com/availproject/op-evm/pkg/test"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashicorp/go-hclog"
	"github.com/test-go/testify/assert"
)

func Test_Builder_Construction_FromParentHash(t *testing.T) {
	executor, bchain, err := test.NewBlockchain(staking.NewVerifier(new(staking.DumbActiveParticipants), hclog.Default()), getGenesisBasePath())
	if err != nil {
		t.Fatal(err)
	}

	h := bchain.Genesis()

	bbf := block.NewBlockBuilderFactory(bchain, executor, hclog.Default())
	if _, err := bbf.FromParentHash(h); err != nil {
		t.Fatal(err)
	}
}

func Test_Builder_Construction_FromBlockchainHead(t *testing.T) {
	executor, bchain, err := test.NewBlockchain(staking.NewVerifier(new(staking.DumbActiveParticipants), hclog.Default()), getGenesisBasePath())
	if err != nil {
		t.Fatal(err)
	}

	bbf := block.NewBlockBuilderFactory(bchain, executor, hclog.Default())
	if _, err := bbf.FromBlockchainHead(); err != nil {
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
	executor, bchain, err := test.NewBlockchain(staking.NewVerifier(new(staking.DumbActiveParticipants), hclog.Default()), getGenesisBasePath())
	if err != nil {
		t.Fatal(err)
	}

	address, privateKey := test.NewAccount(t)
	address2, _ := test.NewAccount(t)

	// Deposit 100 ETH to first account.
	test.DepositBalance(t, address, big.NewInt(0).Mul(big.NewInt(100), common.ETH), bchain, executor)

	// Construct block.Builder w/ the blockchain instance that contains
	// balance for our test account.
	bbf := block.NewBlockBuilderFactory(bchain, executor, hclog.Default())
	bb, err := bbf.FromParentHash(bchain.Header().Hash)
	if err != nil {
		t.Fatal(err)
	}

	amount := big.NewInt(0).Mul(big.NewInt(10), common.ETH)

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

	executor, bchain, err := test.NewBlockchain(staking.NewVerifier(new(staking.DumbActiveParticipants), hclog.Default()), getGenesisBasePath())
	if err != nil {
		t.Fatal(err)
	}

	h := bchain.Genesis()

	bbf := block.NewBlockBuilderFactory(bchain, executor, hclog.Default())
	bb, err := bbf.FromParentHash(h)
	if err != nil {
		t.Fatal(err)
	}

	return bb
}

func TestGetExtraDataFraudProofTarget(t *testing.T) {
	tAssert := assert.New(t)

	testCases := []struct {
		name                string
		input               types.Hash
		expectedHash        types.Hash
		expectedExistsState bool
		expectedError       error
	}{
		{
			name:                "zero address input",
			input:               types.ZeroHash,
			expectedHash:        types.ZeroHash,
			expectedExistsState: false,
		},
		{
			name:                "correct hash input",
			input:               types.StringToHash("1234567890"),
			expectedHash:        types.StringToHash("1234567890"),
			expectedExistsState: true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d: %s", i, tc.name), func(t *testing.T) {
			key := keystore.NewKeyForDirectICAP(rand.Reader)
			blk, err := newBlockBuilder(t).SetExtraDataField(block.KeyFraudProofOf, tc.input.Bytes()).SignWith(key.PrivateKey).Build()
			tAssert.NoError(err)

			hashValue, exists := block.GetExtraDataFraudProofTarget(blk.Header)

			tAssert.Equal(tc.expectedExistsState, exists)
			tAssert.Equal(tc.expectedHash.Bytes(), hashValue.Bytes())
		})
	}
}
