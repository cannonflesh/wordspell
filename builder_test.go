package wordspell

import (
	"io"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/cannonflesh/wordspell/components/bloomfilter"
	"github.com/cannonflesh/wordspell/components/index"
	"github.com/cannonflesh/wordspell/components/langdetect"
	"github.com/cannonflesh/wordspell/components/trademarkindex"
	"github.com/cannonflesh/wordspell/options"
	"github.com/cannonflesh/wordspell/testdata"
)

func TestBuilder_Build(t *testing.T) {
	l, lbuf := testdata.NewTestLogger()

	opt := &options.Options{
		Bloom: bloomfilter.Options{FalsePositiveRate: 0.01},
	}

	itemNames, itemDesc, catNames, tms, err := testdata.CatalogData()
	require.NoError(t, err)

	langs := langdetect.New()

	ruIdxRC, enIdxRC := testdata.IndexData()

	idxSrc := index.NewMockDataSource(t)
	idxSrc.EXPECT().ItemData(0, 100000).
		Return(itemNames, itemDesc, nil).
		Once()
	idxSrc.EXPECT().CategoryNames(0, 10000).
		Return(catNames, nil).
		Once()

	idxStore := index.NewMockDataStore(t)
	idxStore.EXPECT().Save("ru.index", mock.Anything).
		Run(func(_ string, payload io.Reader) {
			cont, err := io.ReadAll(payload)
			require.NoError(t, err)
			contStr := string(cont)
			require.Contains(t, contStr, "лет	24")
			require.Contains(t, contStr, "поможет	24")
		}).
		Return(nil).
		Once()
	idxStore.EXPECT().Save("en.index", mock.Anything).
		Run(func(_ string, payload io.Reader) {
			cont, err := io.ReadAll(payload)
			require.NoError(t, err)
			contStr := string(cont)
			require.Contains(t, contStr, "the	22")
			require.Contains(t, contStr, "orient	15")
		}).
		Return(nil).
		Once()

	tmSrc := trademarkindex.NewMockDataSource(t)
	tmSrc.EXPECT().TradeMarkNames(0, 5000).
		Return(tms, nil).
		Once()

	tmStore := trademarkindex.NewMockDataStore(t)
	tmStore.EXPECT().Save("trademark.index", mock.Anything).
		Run(func(_ string, payload io.Reader) {
			cont, err := io.ReadAll(payload)
			require.NoError(t, err)

			contStr := string(cont)
			require.Contains(t, contStr, "Goodtyre")
			require.Contains(t, contStr, "Revlon Professional")
		}).
		Return(nil).
		Once()

	bloomStore := bloomfilter.NewMockDataStore(t)
	bloomStore.EXPECT().Save("bloom.dat", mock.Anything).
		Run(func(_ string, payload io.Reader) {
			cont, err := io.ReadAll(payload)
			require.NoError(t, err)
			require.Len(t, cont, 456)
		}).
		Return(nil).
		Once()
	bloomStore.EXPECT().DataReader("ru.index").
		Return(ruIdxRC, nil).
		Once()
	bloomStore.EXPECT().DataReader("en.index").
		Return(enIdxRC, nil).
		Once()

	b := &Builder{
		opt:                   opt,
		indexBuilder:          index.NewBuilder(idxSrc, idxStore, langs, l),
		tradeMarkIndexBuilder: trademarkindex.NewBuilder(tmSrc, tmStore, l),
		store:                 bloomStore,
		logger:                l,
	}

	err = b.Build()
	require.NoError(t, err)

	logStr := lbuf.String()
	require.Contains(t, logStr, `[ITEM INDEX BUILD] total names: 100`)
	require.Contains(t, logStr, `[CATEGORY INDEX BUILD] total: 100`)
	require.Contains(t, logStr, `[LANG INDEX SAVE] saving index, lang: en`)
	require.Contains(t, logStr, `[LANG INDEX SAVE] saving index, lang: ru`)
	require.Contains(t, logStr, `[TRADEMARK INDEX BUILD] trademarks total: 100`)
	require.Contains(t, logStr, `[TRADEMARK INDEX SAVE] saved`)
	require.Contains(t, logStr, `[BLOOM FILTER BUILD] built`)
	require.Contains(t, logStr, `[BLOOM FILTER SAVE] saved`)
}
