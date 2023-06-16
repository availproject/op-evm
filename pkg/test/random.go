package test

import (
	"flag"
	"math/rand"
	"testing"
	"time"
)

var (
	// Seed is a global variable used in functions that generate random data.
	// It's value can be specified via a command-line flag `-seed`.
	// By default, it uses the current Unix time.
	Seed = flag.Int64("seed", time.Now().Unix(), "random seed used in tests")
)

// RandomBytes generates a slice of random bytes of the specified size.
// This function uses the global Seed variable for random number generation.
// It's used to produce deterministic results when Seed is specified.
//
// t is a pointer to testing.T, which is the parallel testing interface.
// size is the number of random bytes to generate.
//
// Returns a slice of random bytes.
//
// Example usage:
//
//	func TestRandomBytes(t *testing.T) {
//		bytes := RandomBytes(t, 10)
//		// bytes now holds a slice of 10 random bytes
//	}
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
