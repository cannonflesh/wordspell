package wordspell

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cannonflesh/wordspell/components/bloomfilter"
	"github.com/cannonflesh/wordspell/components/index"
	"github.com/cannonflesh/wordspell/components/langdetect"
	"github.com/cannonflesh/wordspell/components/trademarkindex"
	"github.com/cannonflesh/wordspell/components/wordmutate"
	"github.com/cannonflesh/wordspell/domain"
	"github.com/cannonflesh/wordspell/options"
	"github.com/cannonflesh/wordspell/processors/dimensions"
	"github.com/cannonflesh/wordspell/processors/dimsuffix"
	"github.com/cannonflesh/wordspell/processors/dupremove"
	"github.com/cannonflesh/wordspell/processors/papersizes"
	"github.com/cannonflesh/wordspell/processors/trademarks"
	"github.com/cannonflesh/wordspell/processors/units"
	"github.com/cannonflesh/wordspell/repo/file"
	"github.com/cannonflesh/wordspell/testdata"
)

func goldenSpeller(t *testing.T) (*Service, *testdata.Buffer) {
	lgr, lbuf := testdata.NewTestLogger()

	idxStore := file.New(&file.Options{DataDir: testdata.ThisDir()})
	file.New(&file.Options{DataDir: testdata.ThisDir()})

	idx, err := index.NewService(&options.Options{}, langdetect.New(), idxStore, lgr)
	require.NoError(t, err)

	tmStore := file.New(&file.Options{DataDir: testdata.ThisDir()})
	tm, err := trademarkindex.NewService(tmStore, lgr)
	require.NoError(t, err)

	store := bloomfilter.NewMockDataStore(t)

	preProcessors := []processor{
		trademarks.New(tm),
		dimsuffix.New(),
		dimensions.New(),
		papersizes.New(),
		units.New(),
	}

	postProcessors := []processor{
		dupremove.New(),
	}

	s := &Service{
		langs:  langdetect.New(),
		index:  idx,
		mutate: wordmutate.New(),
		bloom:  bloomfilter.New(&bloomfilter.Options{}, store, lgr),

		preProcessors:  preProcessors,
		postProcessors: postProcessors,

		logger: lgr.WithField(domain.CategoryFieldName, "service.word_speller"),
	}

	err = fillBloomFilter(s.bloom, s.index, s.mutate)
	require.NoError(t, err)

	return s, lbuf
}

func TestService_correctWord(t *testing.T) {
	s, lbuf := goldenSpeller(t)

	t.Run("SuccessShortEn", func(t *testing.T) {
		correct := s.correctWord("1thф")
		require.Equal(t, domain.NewDigestReady("the"), correct)
	})

	t.Run("SuccessLongEn", func(t *testing.T) {
		correct := s.correctWord("internati-nalizфtion")
		require.Equal(t, domain.NewDigestReady("internationalization"), correct)
	})

	t.Run("SuccessShortRu", func(t *testing.T) {
		correct := s.correctWord("ящиг")
		require.Equal(t, domain.NewDigestReady("ящик"), correct)

		correct = s.correctWord("длf")
		require.Equal(t, domain.NewDigestReady("для"), correct)

		correct = s.correctWord("д1я")
		require.Equal(t, domain.NewDigestReady("для"), correct)
	})

	t.Run("SuccessLongRu", func(t *testing.T) {
		correct := s.correctWord("безупасност2")
		require.Equal(t, domain.NewDigestReady("безопасности"), correct)
	})

	t.Run("NoCheckEn", func(t *testing.T) {
		correct := s.correctWord("internationalization")
		require.Equal(t, domain.NewDigestReady("internationalization"), correct)
	})

	t.Run("OneExtraRuneEn", func(t *testing.T) {
		correct := s.correctWord("internationallization")
		require.Equal(t, domain.NewDigestReady("internationalization"), correct)
	})

	t.Run("TwoExtraRunesEn", func(t *testing.T) {
		correct := s.correctWord("interniationallization")
		require.Equal(t, domain.NewDigestReady("internationalization"), correct)
	})

	t.Run("NoCheckRu", func(t *testing.T) {
		correct := s.correctWord("организация")
		require.Equal(t, domain.NewDigestReady("организация"), correct)
	})

	t.Run("OneExtraRuneRu", func(t *testing.T) {
		correct := s.correctWord("организацияя")
		require.Equal(t, domain.NewDigestReady("организация"), correct)
	})

	t.Run("TwoExtraRunesRu", func(t *testing.T) {
		correct := s.correctWord("организзацияя")
		require.Equal(t, domain.NewDigestReady("организация"), correct)
	})

	t.Run("NoCheckNum", func(t *testing.T) {
		correct := s.correctWord("1000.345")
		require.Equal(t, domain.NewDigestReady("1000.345"), correct)
	})

	t.Run("NotInIndex", func(t *testing.T) {
		correct := s.correctWord("really-not-found")
		require.Equal(t, domain.NewDigestRaw("really-not-found"), correct)
	})

	require.Empty(t, lbuf)
}

func TestService_checkWordPairs(t *testing.T) {
	s, _ := goldenSpeller(t)

	t.Run("NothingButPairToCorrect", func(t *testing.T) {
		req := []string{"органи", "зация"}
		digest := domain.ParseDigest(req)

		res := s.checkWordPairs(digest)
		require.Equal(t, `(domain.DigestReady):"организация"`, serializeDigest(res))
	})
	t.Run("NumLangCodeWordBetweeenPairElements", func(t *testing.T) {
		req := []string{"органи", "@International#Business#Machines", "зация"}
		digest := domain.ParseDigest(req)

		res := s.checkWordPairs(digest)
		require.Equal(
			t,
			`(domain.DigestRaw):"органи"|(domain.DigestReady):"International Business Machines"|(domain.DigestRaw):"зация"`,
			serializeDigest(res),
		)
	})
}

func TestService_Correct(t *testing.T) {
	s, _ := goldenSpeller(t)

	enIdx := map[string]uint32{
		"crux":         1000,
		"pux":          1000,
		"fux":          1000,
		"inturnationa": 1000,
		"onewodnami":   1000,
	}

	ruIdx := map[string]uint32{
		"крукс":       1000,
		"пукс":        1000,
		"фукс":        1000,
		"пречистый":   1000,
		"организация": 1000,
		"игрушка для": 1000,
	}

	s.index.SetLangIndex(domain.EnLangCode, enIdx)
	s.index.SetLangIndex(domain.RuLangCode, ruIdx)

	err := fillBloomFilter(s.bloom, s.index, s.mutate)
	require.NoError(t, err)

	t.Run("InternalAsIndividualWord", func(t *testing.T) {
		correct := s.Correct("crax International пакс")
		require.Equal(t, "crux inturnationa пукс", correct)
	})
	t.Run("InternalAsHeadOfValidTradeMark", func(t *testing.T) {
		correct := s.Correct("l = 56cm crax International Business Machines d56ft пакс")
		require.Equal(t, "l=56 cm crux International Business Machines d56 ft пукс", correct)
	})
	t.Run("InternalAsHeadOfIncompleteTradeMark", func(t *testing.T) {
		correct := s.Correct("crax International Machines пакс")
		require.Equal(t, "crux inturnationa Machines пукс", correct)
	})
	t.Run("InternalAsHeadOfIncompleteTradeMark", func(t *testing.T) {
		correct := s.Correct("crax International Machines пакс")
		require.Equal(t, "crux inturnationa Machines пукс", correct)
	})
	t.Run("TwoTrademarks", func(t *testing.T) {
		correct := s.Correct("crax International B.System Of Suncity пакс OneWordName факс")
		require.Equal(t, "crux International B.System Of Suncity пукс OneWordName фукс", correct)
	})
	t.Run("TwoTrademarksAndPartedWord", func(t *testing.T) {
		correct := s.Correct("crax International B.System Of Suncity пре чистый OneWordName факс")
		require.Equal(t, "crux International B.System Of Suncity пречистый OneWordName фукс", correct)
	})
	t.Run("CaseErrorInTheSecondTrademarks", func(t *testing.T) {
		correct := s.Correct("crax International B.System Of Suncity пакс OnewordName факс")
		require.Equal(t, "crux International B.System Of Suncity пукс onewodnami фукс", correct)
	})
	t.Run("NothingButPairToCorrect", func(t *testing.T) {
		correct := s.Correct("органи зация")
		require.Equal(t, "организация", correct)
	})
	t.Run("FindFusedWordWithMistake", func(t *testing.T) {
		correct := s.Correct("игрушкадля")
		require.Equal(t, "игрушка для", correct)
	})
	t.Run("FindCorrectFusedWord", func(t *testing.T) {
		correct := s.Correct("игрушкадля")
		require.Equal(t, "игрушка для", correct)
	})
	t.Run("DupRemove", func(t *testing.T) {
		correct := s.Correct("one two two two three-four four two two two five five-six six-seven eight eight")
		require.Equal(t, "one two three-four fux two five-six six-seven eight", correct)
	})
}

func serializeDigest(dig domain.Digest) string {
	res := make([]string, 0, len(dig))
	for _, v := range dig {
		res = append(res, fmt.Sprintf(`(%T):"%s"`, v, v))
	}

	return strings.Join(res, "|")
}
