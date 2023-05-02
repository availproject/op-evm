package tests

import (
	"crypto/rand"
	"testing"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/andybalholm/brotli"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/hashicorp/go-hclog"
	"github.com/klauspost/compress/zstd"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
	"github.com/maticnetwork/avail-settlement/pkg/test"
	"github.com/test-go/testify/assert"
)

func newBenchmarkBlockBuilder(b *testing.B) block.Builder {
	b.Helper()
	log := hclog.Default()
	log.SetLevel(hclog.Error)
	executor, bchain, err := test.NewBlockchain(staking.NewVerifier(new(staking.DumbActiveParticipants), log), getGenesisBasePath())
	if err != nil {
		b.Fatal(err)
	}

	h := bchain.Genesis()

	bbf := block.NewBlockBuilderFactory(bchain, executor, log)
	bb, err := bbf.FromParentHash(h)
	if err != nil {
		b.Fatal(err)
	}

	return bb
}

func BenchmarkZstdEncoder(b *testing.B) {
	tAssert := assert.New(b)
	key := keystore.NewKeyForDirectICAP(rand.Reader)
	encoder, err := zstd.NewWriter(nil)
	tAssert.NoError(err)

	for i := 0; i < b.N; i++ {
		blk, err := newBenchmarkBlockBuilder(b).SignWith(key.PrivateKey).Build()
		tAssert.NoError(err)
		blkBytes := blk.MarshalRLP()
		encoder.EncodeAll(blkBytes, make([]byte, 0, len(blkBytes)))
	}
}

func BenchmarkZstdDecoder(b *testing.B) {
	tAssert := assert.New(b)
	key := keystore.NewKeyForDirectICAP(rand.Reader)
	encoder, err := zstd.NewWriter(nil)
	tAssert.NoError(err)
	decoder, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))
	tAssert.NoError(err)

	for i := 0; i < b.N; i++ {
		blk, err := newBenchmarkBlockBuilder(b).SignWith(key.PrivateKey).Build()
		tAssert.NoError(err)
		blkBytes := blk.MarshalRLP()
		encBytes := encoder.EncodeAll(blkBytes, make([]byte, 0, len(blkBytes)))
		decBytes, err := decoder.DecodeAll(encBytes, make([]byte, 0, len(encBytes)))
		tAssert.NoError(err)
		dblk := types.Block{}
		err = dblk.UnmarshalRLP(decBytes)
		tAssert.NoError(err)
		tAssert.Equal(blk.Hash().Bytes(), dblk.Hash().Bytes())
	}
}

func BenchmarkBrotliEncoder(b *testing.B) {
	tAssert := assert.New(b)
	key := keystore.NewKeyForDirectICAP(rand.Reader)
	encoder := brotli.NewWriterOptions(nil, brotli.WriterOptions{Quality: 6, LGWin: 0})

	for i := 0; i < b.N; i++ {
		blk, err := newBenchmarkBlockBuilder(b).SignWith(key.PrivateKey).Build()
		tAssert.NoError(err)
		blkBytes := blk.MarshalRLP()
		_, err = encoder.Write(blkBytes)
		tAssert.NoError(err)
		err = encoder.Flush()
		tAssert.NoError(err)
	}

	//encoder.Close()
}
