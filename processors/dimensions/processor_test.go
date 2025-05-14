package dimensions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessor_Process(t *testing.T) {
	proc := New()

	t.Run("OnePattern", func(t *testing.T) {
		words := []string{
			"head",
			"333",
			"x",
			"44Х55",
			"*",
			"777/99кМ",
			"tail",
		}

		check := []string{
			"head",
			"@333*44*55*777*99#км",
			"tail",
		}

		res := proc.Process(words)
		require.Equal(t, check, res)
	})
	t.Run("TwoPatternsTogether", func(t *testing.T) {
		words := []string{
			"head",
			"333",
			"x",
			"44Х55",
			"*",
			"777/99км",
			"15/19Х16",
			"tail",
		}

		check := []string{
			"head",
			"@333*44*55*777*99#км",
			"@15*19*16",
			"tail",
		}

		res := proc.Process(words)
		require.Equal(t, check, res)
	})
	t.Run("PatternAtStart", func(t *testing.T) {
		words := []string{
			"333",
			"x",
			"44Х55",
			"*",
			"777/99км",
			"tail",
		}

		check := []string{
			"@333*44*55*777*99#км",
			"tail",
		}

		res := proc.Process(words)
		require.Equal(t, check, res)
	})
	t.Run("PatternAtEnd", func(t *testing.T) {
		words := []string{
			"head",
			"333",
			"x",
			"44Х55",
			"*",
			"777/99км",
		}

		check := []string{
			"head",
			"@333*44*55*777*99#км",
		}

		res := proc.Process(words)
		require.Equal(t, check, res)
	})
	t.Run("PatternsOnly", func(t *testing.T) {
		words := []string{
			"333",
			"x",
			"44Х55",
			"*",
			"777/99Км",
			"15/19Х16",
			"40*40",
		}

		check := []string{
			"@333*44*55*777*99#км",
			"@15*19*16",
			"@40*40",
		}

		res := proc.Process(words)
		require.Equal(t, check, res)
	})
}
