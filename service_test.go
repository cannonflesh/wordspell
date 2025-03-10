package wordspell

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cannonflesh/wordspell/components/bloomfilter"
	"github.com/cannonflesh/wordspell/components/index"
	"github.com/cannonflesh/wordspell/components/wordmutate"
	"github.com/cannonflesh/wordspell/domain"
	"github.com/cannonflesh/wordspell/options"
	"github.com/cannonflesh/wordspell/testdata"
)

func TestService_Correct(t *testing.T) {
	opt := &options.Options{
		DataDir: "./testdata",
	}

	lgr, lbuf := testdata.NewTestLogger()
	s := New(opt, lgr)

	enIdx := map[string]uint32{
		"the":                  1000,
		"internationalization": 1000,
	}

	ruIdx := map[string]uint32{
		"ящик":         1000,
		"для":          1000,
		"безопасности": 1000,
		"организация":  1000,
	}

	s.index.SetLangIndex(domain.EnLangCode, enIdx)
	s.index.SetLangIndex(domain.RuLangCode, ruIdx)

	fillBloomFilter(s.bloom, s.index, s.mutate, t)

	t.Run("SuccessShortEn", func(t *testing.T) {
		correct := s.Correct("1thф")
		require.Equal(t, "the", correct)
	})

	t.Run("SuccessLongEn", func(t *testing.T) {
		correct := s.Correct("internati-nalizфtion")
		require.Equal(t, "internationalization", correct)
	})

	t.Run("SuccessShortRu", func(t *testing.T) {
		correct := s.Correct("ящиг")
		require.Equal(t, "ящик", correct)

		correct = s.Correct("длf")
		require.Equal(t, "для", correct)

		correct = s.Correct("д1я")
		require.Equal(t, "для", correct)
	})

	t.Run("SuccessLongRu", func(t *testing.T) {
		correct := s.Correct("безупасност2")
		require.Equal(t, "безопасности", correct)
	})

	t.Run("NoCheckEn", func(t *testing.T) {
		correct := s.Correct("internationalization")
		require.Equal(t, "internationalization", correct)
	})

	t.Run("NoCheckRu", func(t *testing.T) {
		correct := s.Correct("организация")
		require.Equal(t, "организация", correct)
	})

	t.Run("NoCheckNum", func(t *testing.T) {
		correct := s.Correct("1000.345")
		require.Equal(t, "1000.345", correct)
	})

	require.Contains(t, lbuf.String(), `category=component.speller_index`)
	require.Contains(t, lbuf.String(), `msg="getting weight: language not detected"`)
}

func fillBloomFilter(bFilter *bloomfilter.Component, idx *index.Component, mutate *wordmutate.Component, t *testing.T) {
	bFilterSizeRu, err := idx.DeletesEstimated(domain.RuLangCode)
	require.NoError(t, err)

	bFilterSizeEn, err := idx.DeletesEstimated(domain.EnLangCode)
	require.NoError(t, err)

	bFilter.Reset(bFilterSizeRu + bFilterSizeEn)

	ruWords, err := idx.Words(domain.RuLangCode)
	require.NoError(t, err)

	enWords, err := idx.Words(domain.EnLangCode)
	require.NoError(t, err)

	for w := range ruWords {
		dts := mutate.Deletes(w)
		bFilter.Add(dts...)
	}

	for w := range enWords {
		dts := mutate.Deletes(w)
		bFilter.Add(dts...)
	}
}
