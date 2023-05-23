package main

import (
	"fmt"
	"os"

	"github.com/maticnetwork/avail-settlement/cmd/availaccount"
	"github.com/maticnetwork/avail-settlement/cmd/keypair"
	"github.com/maticnetwork/avail-settlement/cmd/server"
)

func main() {
	fn := server.Main

	if len(os.Args) > 1 && len(os.Args[1]) > 0 && os.Args[1][0] != '-' {
		switch os.Args[1] {
		case "server":
			fn = server.Main
		case "availaccount":
			fn = availaccount.Main
		case "keypair":
			fn = keypair.Main
		default:
			fmt.Fprintf(os.Stderr, "unknown command: %q\n", os.Args[1])
			return
		}

		// remove the command from args list.
		os.Args = append(os.Args[:0], os.Args[1:]...)
	}

	fn()
}
