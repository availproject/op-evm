package main

import "github.com/ethereum/go-ethereum/ethclient"

func getSequencerClient() (*ethclient.Client, error) {
	return ethclient.Dial(SequencerAddr)
}
