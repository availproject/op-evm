package avail

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/polygon-edge/consensus"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/maticnetwork/avail-settlement/contracts/staking"
	stakingHelper "github.com/maticnetwork/avail-settlement/pkg/staking"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
)

//nolint:golint,unused
func (d *Avail) buildBlock(minerKeystore *keystore.KeyStore, minerAccount accounts.Account, minerPK *keystore.Key, parent *types.Header) (*types.Block, error) {
	header := &types.Header{
		ParentHash: parent.Hash,
		Number:     parent.Number + 1,
		Miner:      minerAccount.Address.Bytes(),
		Nonce:      types.Nonce{},
		GasLimit:   parent.GasLimit, // Inherit from parent for now, will need to adjust dynamically later.
		Timestamp:  uint64(time.Now().Unix()),
	}

	// calculate gas limit based on parent header
	gasLimit, err := d.blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		return nil, err
	}

	header.GasLimit = gasLimit

	// set the timestamp
	parentTime := time.Unix(int64(parent.Timestamp), 0)
	headerTime := parentTime.Add(d.blockTime)

	if headerTime.Before(time.Now()) {
		headerTime = time.Now()
	}

	header.Timestamp = uint64(headerTime.Unix())

	// we need to include in the extra field the current set of validators
	assignExtraValidators(header, ValidatorSet{types.StringToAddress(minerAccount.Address.Hex())})

	d.logger.Info("PARENT BLOCK", "hash", parent.Hash)
	transition, err := d.executor.BeginTxn(parent.StateRoot, header, types.StringToAddress(minerAccount.Address.Hex()))
	if err != nil {
		d.logger.Info("FAILING HERE? 3")
		return nil, err
	}

	stakingAccount, predeployErr := stakingHelper.PredeployStakingSC(
		[]types.Address{types.StringToAddress(minerAccount.Address.Hex())},
		stakingHelper.PredeployParams{
			MinSequencerCount: 1,
			MaxSequencerCount: 10,
		})
	if predeployErr != nil {
		return nil, err
	}

	cstate := d.executor.State()
	snap := cstate.NewSnapshot()
	txn := state.NewTxn(cstate, snap)

	txn.AddBalance(stakingHelper.AddrStakingContract, stakingAccount.Balance)
	txn.SetNonce(stakingHelper.AddrStakingContract, stakingAccount.Nonce)
	txn.SetCode(stakingHelper.AddrStakingContract, stakingAccount.Code)

	for key, value := range stakingAccount.Storage {
		txn.SetState(stakingHelper.AddrStakingContract, key, value)
	}

	if err := transition.SetAccountDirectly(stakingHelper.AddrStakingContract, stakingAccount); err != nil {
		return nil, err
	}

	txns := []*types.Transaction{}

	stakeErr := Stake(transition, gasLimit, types.StringToAddress(minerAccount.Address.Hex()))
	if stakeErr != nil {
		d.logger.Error("failed to query sequencers", "err", stakeErr)
		return nil, stakeErr
	}

	ptxs, err := d.processTxns(gasLimit, transition, txns)
	if err != nil {
		return nil, err
	}

	// Commit the changes
	_, root := transition.Commit()

	// Update the header
	header.StateRoot = root
	header.GasUsed = transition.TotalGas()

	// Build the actual block
	// The header hash is computed inside buildBlock
	block := consensus.BuildBlock(consensus.BuildBlockParams{
		Header:   header,
		Txns:     ptxs,
		Receipts: transition.Receipts(),
	})

	// write the seal of the block after all the fields are completed
	header, err = writeSeal(minerPK.PrivateKey, block.Header)
	if err != nil {
		d.logger.Info("FAILING HERE? 5")
		return nil, err
	}

	block.Header = header

	// compute the hash, this is only a provisional hash since the final one
	// is sealed after all the committed seals
	block.Header.ComputeHash()

	//if err := d.sendBlockToAvail(block); err != nil {
	//	d.logger.Info("FAILING HERE? 6")
	//	return nil, err
	//}

	// Write the block to the blockchain
	if err := d.blockchain.WriteBlock(block, "heck-do-i-know-yet-what-this-is"); err != nil {
		d.logger.Info("FAILING HERE? 7")
		return nil, err
	}

	// after the block has been written we reset the txpool so that
	// the old transactions are removed
	d.txpool.ResetWithHeaders(block.Header)

	addrs, err := QuerySequencers(transition, 5213615, types.StringToAddress(minerAccount.Address.Hex()))
	if err != nil {
		d.logger.Error("failed to query sequencers", "err", err)
		return nil, err
	}
	fmt.Printf("Contract sequencer addresses: %v\n", addrs)

	fmt.Printf("Written block information: %+v\n", block.Header)
	fmt.Printf("Written block transactions: %d\n", len(block.Transactions))

	return block, nil
}

//nolint:golint,unused
func (d *Avail) processTxns(gasLimit uint64, txn *state.Transition, txs []*types.Transaction) ([]*types.Transaction, error) {
	var successful []*types.Transaction

	for _, t := range txs {
		if t.ExceedsBlockGasLimit(gasLimit) {
			if err := txn.WriteFailedReceipt(t); err != nil {
				d.logger.Error("failure to process staking contract transactions - receipt write", "err", err)
				continue
			}

			continue
		}

		if err := txn.Write(t); err != nil {
			d.logger.Error("failure to process staking contract transactions - write", "err", err)
			continue
		}

		successful = append(successful, t)
	}

	return successful, nil
}

func QuerySequencers(t *state.Transition, gasLimit uint64, from types.Address) ([]types.Address, error) {
	method, ok := abi.MustNewABI(staking.StakingABI).Methods["CurrentSequencers"]
	if !ok {
		return nil, errors.New("sequencers method doesn't exist in Staking contract ABI")
	}

	selector := method.ID()
	res, err := t.Apply(&types.Transaction{
		From:     from,
		To:       &stakingHelper.AddrStakingContract,
		Value:    big.NewInt(0),
		Input:    selector,
		GasPrice: big.NewInt(0),
		Gas:      gasLimit,
		Nonce:    t.GetNonce(from),
	})

	if err != nil {
		return nil, err
	}

	if res.Failed() {
		return nil, res.Err
	}

	fmt.Printf("RETURNED VALIDATORS: %+v - err: %v", res, err)
	//return []types.Address{}, nil
	return DecodeValidators(method, res.ReturnValue)
}

func IsSequencer(t *state.Transition, gasLimit uint64, from types.Address) ([]types.Address, error) {
	method, ok := abi.MustNewABI(staking.StakingABI).Methods["CurrentSequencers"]
	if !ok {
		return nil, errors.New("sequencers method doesn't exist in Staking contract ABI")
	}

	selector := method.ID()
	res, err := t.Apply(&types.Transaction{
		From:     from,
		To:       &stakingHelper.AddrStakingContract,
		Value:    big.NewInt(0),
		Input:    selector,
		GasPrice: big.NewInt(0),
		Gas:      gasLimit,
		Nonce:    t.GetNonce(from),
	})

	if err != nil {
		return nil, err
	}

	if res.Failed() {
		return nil, res.Err
	}

	fmt.Printf("RETURNED VALIDATORS: %+v - err: %v", res, err)
	//return []types.Address{}, nil
	return DecodeValidators(method, res.ReturnValue)
}

func Stake(t *state.Transition, gasLimit uint64, from types.Address) error {
	method, ok := abi.MustNewABI(staking.StakingABI).Methods["stake"]
	if !ok {
		return errors.New("sequencers method doesn't exist in Staking contract ABI")
	}

	selector := method.ID()
	res, err := t.Apply(&types.Transaction{
		From:     from,
		To:       &stakingHelper.AddrStakingContract,
		Value:    big.NewInt(0),
		Input:    selector,
		GasPrice: big.NewInt(0),
		Gas:      gasLimit,
		Nonce:    t.GetNonce(from),
	})

	if err != nil {
		return err
	}

	if res.Failed() {
		return res.Err
	}

	fmt.Printf("RETURNED STAKED REQUEST: %+v - err: %v \n", res, err)
	return nil
}

func DecodeValidators(method *abi.Method, returnValue []byte) ([]types.Address, error) {
	decodedResults, err := method.Outputs.Decode(returnValue)
	if err != nil {
		return nil, err
	}

	results, ok := decodedResults.(map[string]interface{})
	if !ok {
		return nil, errors.New("failed type assertion from decodedResults to map")
	}

	web3Addresses, ok := results["0"].([]ethgo.Address)

	if !ok {
		return nil, errors.New("failed type assertion from results[0] to []ethgo.Address")
	}

	addresses := make([]types.Address, 1)
	for idx, waddr := range web3Addresses {
		addresses[idx] = types.Address(waddr)
	}

	return addresses, nil
}
