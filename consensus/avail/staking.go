package avail

import (
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
)

// Checking if the staking contract is deployed, if it is skipping.
// If not deployed, deploying the contract itself.
func (d *Avail) deployStakingContract(minerKeystore *keystore.KeyStore, miner accounts.Account, minerPK *keystore.Key) error {

	// Check if the contract is deployed to a specific address (code)
	// If yes, all good return ok
	// If not:
	// Create new staking address
	// Figure out a way how to create block and transactions inside
	// Push the block to avail
	// Push the block to local blockchain

	cstate := d.executor.State()
	snap := cstate.NewSnapshot()
	txn := state.NewTxn(cstate, snap)

	contractCode := txn.GetCode(staking.AddrStakingContract)

	// Contract code available, we won't redeploy the contract and for now assume
	// that everything is correctly applied
	if len(contractCode) > 0 {
		return nil
	}

	// There is no contract code at associated staking address, we are going to start deploying the contract
	/* 	contractAddr := types.StringToAddress(staking.AddrStakingContract.String())
	   	deployTxn := &types.Transaction{
	   		Nonce: 0,
	   		From:  types.Address(miner.Address),
	   		To:    &contractAddr,
	   	}

	   	if err := d.txpool.AddTx(deployTxn); err != nil {
	   		return err
	   	} */

	header := d.blockchain.Header()
	_, err := d.buildBlock(minerKeystore, miner, minerPK, header)
	if err != nil {
		d.logger.Error("failed to mine block", "err", err)
		return err
	}

	return nil
}
