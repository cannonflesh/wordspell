package dupremove

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProcessor_Process(t *testing.T) {
	p := New()

	check := []string{
		"one",
		"two",
		"two",
		"two",
		"three-four",
		"four",
		"two",
		"two",
		"two",
		"five",
		"five-six",
		"six-seven",
		"eight",
		"eight",
	}

	res := p.Process(check)
	require.Equal(t, []string{"one", "two", "three-four", "two", "five-six", "six-seven", "eight"}, res)
}
