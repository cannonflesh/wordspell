package wordspell

import (
	"github.com/cannonflesh/wordspell/options"
	"github.com/cannonflesh/wordspell/testdata"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestService_Correct(t *testing.T) {
	opt := &options.Options{
		DataDir: "./data",
	}

	lgr, lbuf := testdata.NewTestLogger()
	s := New(opt, lgr)

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

func TestService_Correctness(t *testing.T) {
	opt := &options.Options{
		DataDir: "./data",
	}

	lgr, _ := testdata.NewTestLogger()
	s := New(opt, lgr)

	t.Run("Check", func(t *testing.T) {
		start := time.Now()
		correct := s.Correct("подсигар")
		t.Log(correct)
		t.Log(time.Since(start), "elapsed")
	})
}
