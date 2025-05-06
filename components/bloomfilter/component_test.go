package bloomfilter

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cannonflesh/wordspell/testdata"
)

func TestComponent_All(t *testing.T) {
	lgr, lbuf := testdata.NewTestLogger()
	opt := &Options{FalsePositiveRate: 0.01}

	store := NewMockDataStore(t)

	bf := New(opt, store, lgr)

	found := bf.Test("хрензначо")
	require.False(t, found)

	bf.Reset(100)
	bf.Add("хрензначо")

	found = bf.Test("хрензначо")
	require.True(t, found)

	require.Empty(t, lbuf.String())
}
