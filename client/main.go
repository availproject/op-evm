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
}
