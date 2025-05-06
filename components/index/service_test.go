package index

import (
	"bytes"
	"github.com/cannonflesh/wordspell/components/langdetect"
	"github.com/cannonflesh/wordspell/options"
	"github.com/stretchr/testify/require"
	"io"
	"sync"
	"testing"
)

func TestService_runeLen(t *testing.T) {
	tstr := "№ русский текст"

	require.Equal(t, 29, len(tstr))
	require.Equal(t, uint(15), runeLen(tstr))
}

func TestService_All(t *testing.T) {
	store := NewMockDataStore(t)
	store.EXPECT().DataReader(langCodeIndexKey(ruLangCode)).
		Return(goldenRuData(), nil).
		Once()
	store.EXPECT().DataReader(langCodeIndexKey(enLangCode)).
		Return(goldenEnData(), nil).
		Once()

	s := &Service{
		opt: &options.Options{
			Langs: []langCode{ruLangCode, enLangCode},
		},
		langs: langdetect.New(),
		store: store,
		index: make(wordCollection),
		mu:    sync.RWMutex{},
	}

	err := s.load()
	require.NoError(t, err)

	t.Run("CheckLoad", func(t *testing.T) {
		idx, ok := s.index[ruLangCode]
		require.True(t, ok)
		require.Len(t, idx, 30)
		require.Equal(t, uint32(1703405), idx["цвет"])
		require.Equal(t, uint32(528614), idx["рост"])
		require.Equal(t, uint32(245425), idx["рост цвет"])

		idx, ok = s.index[enLangCode]
		require.True(t, ok)
		require.Len(t, idx, 30)
		require.Equal(t, uint32(159700), idx["in"])
		require.Equal(t, uint32(60747), idx["german"])
		require.Equal(t, uint32(57616), idx["german edition"])
	})
	t.Run("CheckWeight", func(t *testing.T) {
		require.Equal(t, uint32(57616), s.Weight("german edition"))
		require.Equal(t, uint32(157718), s.Weight("edition"))
		require.Equal(t, uint32(1703405), s.Weight("цвет"))
		require.Equal(t, uint32(245425), s.Weight("рост цвет"))
	})
	t.Run("CheckWords", func(t *testing.T) {
		wChan, err := s.Words()
		require.NoError(t, err)
		found := 0
		total := 0
		for w := range wChan {
			if w == "цвет" ||
				w == "книга" ||
				w == "размер цвет" {
				found++
			}

			total++
		}
		require.Equal(t, 3, found)
		require.Equal(t, 60, total)
	})
	t.Run("CheckDeletesEstimates", func(t *testing.T) {
		destCount, err := s.DeletesEstimated()
		require.NoError(t, err)
		require.Equal(t, uint(1668), destCount)
	})
}

func goldenRuData() io.ReadCloser {
	return io.NopCloser(
		bytes.NewBufferString(`и	2959334
для	2192256
в	1968119
цвет	1703405
с	1658600
на	1336862
размер	1105016
книга	1010744
не	726503
из	573814
рост	528614
а	456690
набор	448342
от	349867
или	326270
размер цвет	304685
к	294578
белый	286006
по	285009
микс	278725
чёрный	250540
рост цвет	245425
это	236379
при	217719
вы	189921
комплект	186224
девочки	180776
женская	177730
для девочки	176891
как	171883`))
}

func goldenEnData() io.ReadCloser {
	return io.NopCloser(
		bytes.NewBufferString(`the	630307
of	580453
and	282696
a	256054
of the	194762
name	162006
in	159700
edition	157718
et	148436
volume	120162
to	84605
des	76040
german	60747
on	60142
for	58873
german edition	57616
led	57178
with	55614
und	52201
french	46139
by	46011
from	41661
in the	41463
french edition	41317
history	40850
an	38978
die	38301
les	38297
i	37537
new	36146`))
}
