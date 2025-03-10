package bloomfilter

import (
	"github.com/cannonflesh/wordspell/options"
	"github.com/cannonflesh/wordspell/testdata"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestComponent_All(t *testing.T) {
	t.Run("EmptyBloomFilter", func(t *testing.T) {
		lgr, lbuf := testdata.NewTestLogger()
		opt := &options.Options{DataDir: "./"}
		bf := New(opt, lgr)

		found := bf.Test("хрензначо")
		require.False(t, found)

		bf.Reset(100)
		bf.Add("хрензначо")

		found = bf.Test("хрензначо")
		require.True(t, found)

		require.Contains(t, lbuf.String(), `category=components.bloom_filter`)
		require.Contains(t, lbuf.String(), `msg="loading bloom filter data from file"`)
		require.Contains(t, lbuf.String(), `error="open bloom.dat: no such file or directory"`)
	})
}
