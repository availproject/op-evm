package main

import (
	"log"

	"github.com/maticnetwork/avail-settlement/cmd/availaccount"
)

func main() {
	if err := availaccount.GetCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
