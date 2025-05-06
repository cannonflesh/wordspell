package units

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessor_Process(t *testing.T) {
	proc := New()

	t.Run("ThreePatternsInTheMiddle", func(t *testing.T) {
		words := []string{
			"head",
			"3.14", "mm",
			"spacer",
			"9.14-", "15.33", "-99кг",
			"72,18шт",
			"tail",
		}

		check := []string{
			"head",
			"@3.14#mm",
			"spacer",
			"@9.14-15.33-99#кг",
			"@72,18#шт",
			"tail",
		}

		res := proc.Process(words)
		require.Equal(t, check, res)
	})
	t.Run("PatternAtStart", func(t *testing.T) {
		words := []string{
			"3.14", "mm",
			"spacer",
			"9.14-", "15.33", "-99кг",
			"72,18шт",
			"tail",
		}

		check := []string{
			"@3.14#mm",
			"spacer",
			"@9.14-15.33-99#кг",
			"@72,18#шт",
			"tail",
		}

		res := proc.Process(words)
		require.Equal(t, check, res)
	})
	t.Run("PatternAtEnd", func(t *testing.T) {
		words := []string{
			"head",
			"l 3.14", "mm",
			"spacer",
			"9.14-", "15.33", "-99кг",
			"72,18шт",
		}

		check := []string{
			"head",
			"@l3.14#mm",
			"spacer",
			"@9.14-15.33-99#кг",
			"@72,18#шт",
		}

		res := proc.Process(words)
		require.Equal(t, check, res)
	})
	t.Run("PatternsOnly", func(t *testing.T) {
		words := []string{
			"3.14mm",
			"9.14-", "15.33", "-99кг",
			"d = 72,18", "in",
			"55", "-75", "%",
		}

		check := []string{
			"@3.14#mm",
			"@9.14-15.33-99#кг",
			"@d=72,18#in",
			"@55-75%",
		}

		res := proc.Process(words)
		require.Equal(t, check, res)
	})
}
