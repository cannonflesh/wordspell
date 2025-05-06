package trademarkindex

import (
	"github.com/stretchr/testify/mock"
	"io"
	"testing"

	"github.com/cannonflesh/wordspell/testdata"
	"github.com/stretchr/testify/require"
)

func TestBuilder_LoadIndexData(t *testing.T) {
	l, _ := testdata.NewTestLogger()
	source := NewMockDataSource(t)
	source.EXPECT().TradeMarkNames(0, sourceBatchLen).
		Return(goldenTradeMarks(), nil)
	store := NewMockDataStore(t)
	store.EXPECT().Save(mock.Anything, mock.Anything).Run(func(k string, d io.Reader) {
		require.Equal(t, storeKey, k)
		var payload = make([]byte, 128)
		read, err := d.Read(payload)
		require.NoError(t, err)
		require.Equal(t, 80, read)
	}).Return(nil).
		Once()
	d := NewBuilder(source, store, l)
	err := d.LoadIndexData()
	require.NoError(t, err)
}

func goldenTradeMarks() []string {
	return []string{
		"Mazda",
		"Mazda super",
		"Mazda super puper",
		"Mazda puper duper cooper",
		"Cooper super dooper",
	}
}
