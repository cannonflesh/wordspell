package index

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cannonflesh/wordspell/components/langdetect"
	"github.com/cannonflesh/wordspell/options"
	"github.com/cannonflesh/wordspell/testdata"
)

func TestService_All(t *testing.T) {
	l, lbuf := testdata.NewTestLogger()

	// Файлы данных подбираются из директории testdata,
	// они там урезаны как раз по 1000 записей.
	s := New(goldenOpt(), langdetect.New(), l)

	require.Len(t, s.index, 2)
	require.Len(t, s.index["en"], 1000)
	require.Len(t, s.index["ru"], 1000)

	// Число.
	// Мы не храним числа в словаре и не исправляем опечатки в числах.
	require.Equal(t, uint32(numWeight), s.Weight("12056"))
	require.Equal(t, uint32(numWeight), s.Weight("12056.223"))
	require.Equal(t, uint32(numWeight), s.Weight("12056,223"))

	// А вот это - не число, язык для таких слов мы не распознаем.
	require.Equal(t, uint32(0), s.Weight("12056.22.3"))
	require.Contains(t, lbuf.String(), `getting weight: language not detected`)
	require.Contains(t, lbuf.String(), `category=component.speller_index`)

	// Русский индекс.
	require.Equal(t, uint32(936), s.Weight("военный"))
	// Английский индекс.
	require.Equal(t, uint32(68026), s.Weight("motorola"))

	// Нет в индексе.
	// NB! ответственность за приведение слов к lowerCase - на вызывающей стороне.
	require.Equal(t, uint32(0), s.Weight("MoToRoLa"))
	require.Equal(t, uint32(0), s.Weight("нет-такого-слова"))
}

func goldenOpt() *options.Options {
	return &options.Options{DataDir: "../../testdata"}
}
