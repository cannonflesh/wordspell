package dimsuffix

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProcessor_Process(t *testing.T) {
	proc := New()

	t.Run("OnePattern", func(t *testing.T) {
		req := []string{
			"head",
			"2",
			"d",
			"tail",
		}

		check := []string{
			"head",
			"@2D",
			"tail",
		}

		res := proc.Process(req)
		require.Equal(t, check, res)
	})
	t.Run("TwoPatternsTogether", func(t *testing.T) {
		req := []string{
			"head",
			"2",
			"d",
			"3д",
			"tail",
		}

		check := []string{
			"head",
			"@2D",
			"@3D",
			"tail",
		}

		res := proc.Process(req)
		require.Equal(t, check, res)
	})
	t.Run("PatternAtStart", func(t *testing.T) {
		req := []string{
			"3д",
			"tail",
		}

		check := []string{
			"@3D",
			"tail",
		}

		res := proc.Process(req)
		require.Equal(t, check, res)
	})
	t.Run("TwoPatternAtEnd", func(t *testing.T) {
		req := []string{
			"head",
			"2",
			"d",
		}

		check := []string{
			"head",
			"@2D",
		}

		res := proc.Process(req)
		require.Equal(t, check, res)
	})
	t.Run("PatternsOnly", func(t *testing.T) {
		req := []string{
			"2",
			"d",
			"5 Д",
			"3",
			"D",
		}

		check := []string{
			"@2D",
			"@5D",
			"@3D",
		}

		res := proc.Process(req)
		require.Equal(t, check, res)
	})
}
