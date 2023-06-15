package tests

import (
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/0xPolygon/polygon-edge/consensus"
	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/consensus/avail"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/blockchain"
	"github.com/maticnetwork/avail-settlement/pkg/common"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
	"github.com/maticnetwork/avail-settlement/pkg/test"
)

func Test_ForkBlockchain(t *testing.T) {
	bootstrapSequencer := NewTestNode(t)
	ordinarySequencer := NewTestNode(t)
	watchtower := NewTestNode(t)

	// Deposit 50 ETH to all accounts.
	EnsureBalances(t, big.NewInt(0).Mul(big.NewInt(50), common.ETH), []*TestNode{bootstrapSequencer, ordinarySequencer, watchtower})

	//+------------------------------+
	//| Blockchain network emulation |
	//+------------------------------+
	// 1. Produce 7 blocks by bootstrap sequencer.
	// 2. Produce 7 blocks by ordinary sequencer.
	// 3. Produce 1+1+3 blocks by boostrap sequencer.
	// 3.1. 1st block is correct.
	// 3.2. 2nd block is "malicious".
	// 3.3. 3rd - 6th blocks are correct.
	// 4. Fork just before the "malicious" block by ordinary sequencer
	//    (i.e. emulate a dispute resolution). This *should* cause a re-org
	//    in the chain.
	// 5. Produce 2 blocks by ordinary sequencer.
	// 6. Produce 4 blocks by bootstrap sequencer.
	// 7. Inspect the results.

	//---------------------------------------------------------------------------
	// Step 1. Produce 7 blocks by bootstrap sequencer.
	//---------------------------------------------------------------------------
	for i := 0; i < 7; i++ {
		blk := bootstrapSequencer.ProduceBlock()
		ordinarySequencer.WriteBlockToChain(blk)
		watchtower.WriteBlockToChain(blk)
	}

	//---------------------------------------------------------------------------
	// Step 2. Produce 7 blocks by ordinary sequencer.
	//---------------------------------------------------------------------------
	for i := 0; i < 7; i++ {
		blk := ordinarySequencer.ProduceBlock()
		bootstrapSequencer.WriteBlockToChain(blk)
		watchtower.WriteBlockToChain(blk)
	}

	//---------------------------------------------------------------------------
	// Step 3. Produce 1+1+3 blocks by boostrap sequencer.
	//---------------------------------------------------------------------------
	// Step 3.1. 1st block is correct.
	//---------------------------------------------------------------------------
	forkHeadBlk := bootstrapSequencer.ProduceBlock()
	ordinarySequencer.WriteBlockToChain(forkHeadBlk)
	watchtower.WriteBlockToChain(forkHeadBlk)

	//---------------------------------------------------------------------------
	// Step 3.2. 2nd block is "malicious".
	//---------------------------------------------------------------------------
	//invalidBlk := bootstrapSequencer.ProduceInvalidBlock()
	okayishBlk := bootstrapSequencer.ProduceBlock()
	ordinarySequencer.WriteBlockToChain(okayishBlk)
	watchtower.WriteBlockToChain(okayishBlk)

	//---------------------------------------------------------------------------
	// Step 3.3. 3rd - 6th blocks are correct.
	//---------------------------------------------------------------------------
	for i := 0; i < 3; i++ {
		blk := bootstrapSequencer.ProduceBlock()
		ordinarySequencer.WriteBlockToChain(blk)
		watchtower.WriteBlockToChain(blk)
	}

	//---------------------------------------------------------------------------
	// 4. Fork just before the "malicious" block by ordinary sequencer
	//---------------------------------------------------------------------------
	blk := ordinarySequencer.ProduceForkBlockFrom(forkHeadBlk.Header)
	bootstrapSequencer.WriteBlockToChain(blk)
	watchtower.WriteBlockToChain(blk)

	//---------------------------------------------------------------------------
	// 5. Produce 2 blocks by ordinary sequencer.
	//---------------------------------------------------------------------------
	for i := 0; i < 2; i++ {
		blk = ordinarySequencer.ProduceBlock()
		bootstrapSequencer.WriteBlockToChain(blk)
		watchtower.WriteBlockToChain(blk)
	}

	//---------------------------------------------------------------------------
	// 6. Produce 4 blocks by bootstrap sequencer.
	//---------------------------------------------------------------------------
	for i := 0; i < 4; i++ {
		blk = bootstrapSequencer.ProduceBlock()
		ordinarySequencer.WriteBlockToChain(blk)
		watchtower.WriteBlockToChain(blk)
	}

	//---------------------------------------------------------------------------
	// 7. Inspect the results.
	//---------------------------------------------------------------------------
	hdr := bootstrapSequencer.blockchain.Header()
	t.Logf("bootstrap sequencer Header(): #%d, %q", hdr.Number, hdr.Hash.String())
	hdr = ordinarySequencer.blockchain.Header()
	t.Logf("ordinary sequencer Header(): #%d, %q", hdr.Number, hdr.Hash.String())
	hdr = watchtower.blockchain.Header()
	t.Logf("watchtower Header(): #%d, %q", hdr.Number, hdr.Hash.String())
}

//////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////
/////////////// Test types & support functions  //////////////////////////////
//////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////

type TestNode struct {
	t *testing.T

	blockchain          *blockchain.Blockchain
	executor            *state.Executor
	blockBuilderFactory block.BlockBuilderFactory

	address types.Address
	signKey *ecdsa.PrivateKey
}

func NewTestNode(t *testing.T) *TestNode {
	e, bc, err := test.NewBlockchain(staking.NewVerifier(new(staking.DumbActiveParticipants), hclog.Default()), getGenesisBasePath())
	if err != nil {
		t.Fatal(err)
	}

	addr, signKey := test.NewAccount(t)

	return &TestNode{
		t: t,

		blockchain:          bc,
		executor:            e,
		blockBuilderFactory: block.NewBlockBuilderFactory(bc, e, hclog.Default()),

		address: addr,
		signKey: signKey,
	}
}

func (tc *TestNode) ProduceForkBlockFrom(hdr *types.Header) *types.Block {
	bb, err := tc.blockBuilderFactory.FromParentHash(hdr.Hash)
	if err != nil {
		tc.t.Fatal(err)
	}

	// Transfer some WEI from "FaucetAccount" to node address.
	tx := &types.Transaction{
		From:     tc.address,
		To:       &test.FaucetAccount,
		Value:    big.NewInt(10000000),
		Gas:      100000,
		GasPrice: big.NewInt(1),
	}

	signer := &crypto.FrontierSigner{}
	tx, err = signer.SignTx(tx, tc.signKey)
	if err != nil {
		tc.t.Fatal(err)
	}

	td, found := tc.blockchain.GetChainTD()
	if !found {
		tc.t.Fatal("couldn't find chain TD!")
	}

	blk, err := bb.
		SetBlockNumber(tc.blockchain.Header().Number + 1).
		SetCoinbaseAddress(tc.address).
		SetDifficulty(td.Uint64() + 1).
		AddTransactions(tx).
		SignWith(tc.signKey).
		Build()
	if err != nil {
		tc.t.Fatal(err)
	}

	err = tc.blockchain.WriteBlock(blk, "test")
	if err != nil {
		tc.t.Fatal(err)
	}

	return blk
}

func (tc *TestNode) ProduceBlock() *types.Block {
	return tc.ProduceBlockFrom(tc.blockchain.Header())
}

func (tc *TestNode) ProduceBlockFrom(hdr *types.Header) *types.Block {
	bb, err := tc.blockBuilderFactory.FromParentHash(hdr.Hash)
	if err != nil {
		tc.t.Fatal(err)
	}

	// Transfer some WEI from "FaucetAccount" to node address.
	tx := &types.Transaction{
		From:     tc.address,
		To:       &test.FaucetAccount,
		Value:    big.NewInt(10000000),
		Gas:      100000,
		GasPrice: big.NewInt(1),
	}

	signer := &crypto.FrontierSigner{}
	tx, err = signer.SignTx(tx, tc.signKey)
	if err != nil {
		tc.t.Fatal(err)
	}

	blk, err := bb.
		SetCoinbaseAddress(tc.address).
		AddTransactions(tx).
		SignWith(tc.signKey).
		Build()
	if err != nil {
		tc.t.Fatal(err)
	}

	err = tc.blockchain.WriteBlock(blk, "test")
	if err != nil {
		tc.t.Fatal(err)
	}

	return blk
}

func (tc *TestNode) ProduceInvalidBlock() *types.Block {
	parent := tc.blockchain.Header()

	header := &types.Header{
		ParentHash: parent.Hash,
		Number:     parent.Number + 1,
		Miner:      tc.address.Bytes(),
		Nonce:      types.Nonce{},
		GasLimit:   parent.GasLimit, // Inherit from parent for now, will need to adjust dynamically later.
		Timestamp:  uint64(time.Now().Unix()),
	}

	// calculate gas limit based on parent header
	gasLimit, err := tc.blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		tc.t.Fatal(err)
	}

	header.GasLimit = gasLimit

	// set the timestamp
	parentTime := time.Unix(int64(parent.Timestamp), 0)
	headerTime := parentTime.Add(20 * time.Second)

	if headerTime.Before(time.Now()) {
		headerTime = time.Now()
	}

	header.Timestamp = uint64(headerTime.Unix())

	// we need to include in the extra field the current set of validators
	err = block.AssignExtraValidators(header, avail.ValidatorSet{tc.address})
	if err != nil {
		tc.t.Fatal(err)
	}

	transition, err := tc.executor.BeginTxn(parent.StateRoot, header, tc.address)
	if err != nil {
		tc.t.Fatal(err)
	}

	// Legit tx: Transfer some WEI from "FaucetAccount" to node address.
	tx := &types.Transaction{
		From:     test.FaucetAccount,
		To:       &tc.address,
		Value:    big.NewInt(10000000),
		Gas:      100000,
		GasPrice: big.NewInt(1),
	}

	signer := &crypto.FrontierSigner{}
	tx, err = signer.SignTx(tx, test.FaucetSignKey)
	if err != nil {
		tc.t.Fatal(err)
	}

	err = transition.Write(tx)
	if err != nil {
		tc.t.Fatal(err)
	}

	txns := []*types.Transaction{tx}

	tx = &types.Transaction{
		From:     tc.address,
		To:       &test.FaucetAccount,
		Value:    big.NewInt(100000000000),
		Gas:      10000,
		GasPrice: big.NewInt(1),
	}

	tx, err = signer.SignTx(tx, test.FaucetSignKey)
	if err != nil {
		tc.t.Fatal(err)
	}
	// ^.... that transaction hasn't been executed -> it will yield error when
	//       checking the block.

	txns = append(txns, tx)

	// Commit the changes
	_, root := transition.Commit()

	// Update the header
	header.StateRoot = root
	header.GasUsed = transition.TotalGas()

	// Build the actual block
	// The header hash is computed inside buildBlock
	blk := consensus.BuildBlock(consensus.BuildBlockParams{
		Header:   header,
		Txns:     txns,
		Receipts: transition.Receipts(),
	})

	// write the seal of the block after all the fields are completed
	header, err = block.WriteSeal(tc.signKey, blk.Header)
	if err != nil {
		tc.t.Fatal(err)
	}

	blk.Header = header

	// compute the hash, this is only a provisional hash since the final one
	// is sealed after all the committed seals
	blk.Header.ComputeHash()

	err = tc.blockchain.WriteBlock(blk, "test")
	if err != nil {
		tc.t.Logf("failed to write invalid block to blockchain: %s", err)
	}

	return blk
}

func (tc *TestNode) WriteBlockToChain(b *types.Block) {
	err := tc.blockchain.WriteBlock(b, "foobar")
	if err != nil {
		tc.t.Logf("failed to write block to blockchain: %s", err)
	}
}

func EnsureBalances(t *testing.T, amount *big.Int, nodes []*TestNode) {
	var addrs []types.Address
	for _, n := range nodes {
		addrs = append(addrs, n.address)
	}

	for _, n := range nodes {
		for _, addr := range addrs {
			test.DepositBalance(t, addr, amount, n.blockchain, n.executor)
		}
	}
}