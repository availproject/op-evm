package test

import (
	"flag"
	"math/rand"
	"testing"
	"time"
)

var (
	// Seed used in functions that generate random data. Its value can be
	// specified with -seed parameter. It defaults to time.Now().Unix().
	Seed = flag.Int64("seed", time.Now().Unix(), "random seed used in tests")
)

// RandomBytes generates requested number of random bytes.
func RandomBytes(t *testing.T, size int) []byte {
	t.Helper()

	// This allows deterministic tests when seed is specified.
	rnd := rand.New(rand.NewSource(*Seed))
	buf := make([]byte, size)

	bytesRead := 0
	for bytesRead < size {
		n, err := rnd.Read(buf[bytesRead:])
		if err != nil {
			t.Fatalf("error while generating random bytes: %s", err)
		}

		bytesRead += n
	}

	return buf
}
