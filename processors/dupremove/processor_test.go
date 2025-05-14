package dupremove

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessor_Process(t *testing.T) {
	p := New()

	check := []string{
		"one",
		"two",
		"two",
		"two",
		"three-Four",
		"four",
		"two",
		"Two",
		"two",
		"five",
		"five-six",
		"six-seven",
		"eight",
		"eight",
	}

	res := p.Process(check)
	require.Equal(t, []string{"one", "two", "three-Four", "two", "five-six", "six-seven", "eight"}, res)
}
