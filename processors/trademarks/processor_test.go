package trademarks

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/components/trademarkindex"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/repo/file"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/testdata"
)

func TestProcessor_Process(t *testing.T) {
	lgr, _ := testdata.NewTestLogger()
	tmStore := file.New(&file.Options{DataDir: testdata.ThisDir()})
	tmIdx, err := trademarkindex.NewService(tmStore, lgr)
	require.NoError(t, err)

	proc := New(tmIdx)

	t.Run("AllFound", func(t *testing.T) {
		testReq := []string{
			"crax",
			"International", "Business", "Machines",
			"pax",
			"fax",
			"F+", "special",
			"crax",
			"OneWordName",
			"tail",
			"tail",
		}

		check := []string{
			"crax",
			"@International#Business#Machines",
			"pax",
			"fax",
			"@F+#special",
			"crax",
			"@OneWordName",
			"tail",
			"tail",
		}

		res := proc.Process(testReq)
		require.Len(t, res, 9)
		require.Equal(t, check, res)
	})
	t.Run("NoOneWordInternationalTradeMark", func(t *testing.T) {
		testReq := []string{
			"crax",
			"International", "noname", "Business", "Machines",
			"pax",
		}

		check := []string{
			"crax",
			"International", "noname", "Business", "Machines",
			"pax",
		}

		res := proc.Process(testReq)
		require.Len(t, res, 6)
		require.Equal(t, check, res)
	})
	t.Run("IncompleteTradeMarkName", func(t *testing.T) {
		testReq := []string{
			"crax",
			"International", "Business",
			"pax",
		}

		check := []string{
			"crax",
			"International", "Business",
			"pax",
		}

		res := proc.Process(testReq)
		require.Len(t, res, 4)
		require.Equal(t, check, res)
	})
	t.Run("IncorrectCapitalization", func(t *testing.T) {
		testReq := []string{
			"crax",
			"International", "business", "Machines",
			"pax",
		}

		check := []string{
			"crax",
			"International", "business", "Machines",
			"pax",
		}

		res := proc.Process(testReq)
		require.Len(t, res, 5)
		require.Equal(t, check, res)
	})
}
