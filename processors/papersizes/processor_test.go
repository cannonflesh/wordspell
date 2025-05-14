package papersizes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessor_Process(t *testing.T) {
	proc := New()

	t.Run("OnePaperSize", func(t *testing.T) {
		req := []string{
			"head",
			"a",
			"5",
			"tail",
		}

		check := []string{
			"head",
			"@A5",
			"tail",
		}

		res := proc.Process(req)
		require.Equal(t, check, res)
	})
	t.Run("TwoPaperSizeTogether", func(t *testing.T) {
		req := []string{
			"head",
			"a",
			"5",
			"b",
			"2",
			"tail",
		}

		check := []string{
			"head",
			"@A5",
			"@B2",
			"tail",
		}

		res := proc.Process(req)
		require.Equal(t, check, res)
	})
	t.Run("PaperSizeAtStart", func(t *testing.T) {
		req := []string{
			"a",
			"5",
			"tail",
		}

		check := []string{
			"@A5",
			"tail",
		}

		res := proc.Process(req)
		require.Equal(t, check, res)
	})
	t.Run("PaperSizeAtEnd", func(t *testing.T) {
		req := []string{
			"head",
			"a",
			"5",
		}

		check := []string{
			"head",
			"@A5",
		}

		res := proc.Process(req)
		require.Equal(t, check, res)
	})
	t.Run("PaperSizesOnly", func(t *testing.T) {
		req := []string{
			"a",
			"5",
			"b3",
			"a6",
			"B",
			"1",
		}

		check := []string{
			"@A5",
			"@B3",
			"@A6",
			"@B1",
		}

		res := proc.Process(req)
		require.Equal(t, check, res)
	})
}
