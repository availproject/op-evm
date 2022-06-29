package main

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("http://127.0.0.1:20002")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("client: %#v", client)

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Got the header number: %s", header.Number.String())

	block, err := client.BlockByNumber(context.Background(), header.Number)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Block number: %v", block.Number().Uint64())
	log.Printf("Block time: %v", block.Time())
	log.Printf("Block difficulty: %v", block.Difficulty().Uint64())
	log.Printf("Block hash (hex): %v", block.Hash().Hex())
	log.Printf("Block transactions length: %v", len(block.Transactions()))
}
