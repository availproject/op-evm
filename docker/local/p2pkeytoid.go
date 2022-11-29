package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "error: keyfile parameter missing\nusage: %s <keyfile>\n", os.Args[0])
		os.Exit(1)
	}

	buf, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	buf, err = hex.DecodeString(string(buf))
	if err != nil {
		panic(err)
	}

	libp2pKey, err := crypto.UnmarshalPrivateKey(buf)
	if err != nil {
		panic(err)
	}

	id, err := peer.IDFromPrivateKey(libp2pKey)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", id.String())
}
