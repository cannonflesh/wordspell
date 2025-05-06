package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanTextRE(t *testing.T) {
	require.Equal(t,
		[]string{"Тапочки-зверушки", "Серый", "мишка", ",", "р-р", "38", "Can`t", "ain't"},
		strings.Fields(
			CleanTextRE.ReplaceAllString("Тапочки-зверушки \"Серый мишка\", р-р 38 Can`t ain't\n", " "),
		),
	)
}

func TestCleanIndexRE(t *testing.T) {
	require.Equal(t,
		[]string{"Тапочки-зверушки", "Серый", "мишка", ",", "р-р", "38", "Can`t", "ain't"},
		strings.Fields(
			CleanTextRE.ReplaceAllString("Тапочки-зверушки \"Серый мишка\", р-р 38 Can`t ain't\n", " "),
		),
	)
}
