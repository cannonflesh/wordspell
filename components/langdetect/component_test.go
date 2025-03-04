package langdetect

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComponent_LangByWord(t *testing.T) {
	c := New()

	t.Run("NumberLang", func(t *testing.T) {
		lang := c.LangByWord("12056")
		require.Equal(t, numLangCode, lang)

		lang = c.LangByWord("12056.223")
		require.Equal(t, numLangCode, lang)

		lang = c.LangByWord("12056,223")
		require.Equal(t, numLangCode, lang)

		// Два десятичных символа - не число.
		lang = c.LangByWord("12056.22.3")
		require.Equal(t, unknownLangCode, lang)
	})

	t.Run("Success", func(t *testing.T) {
		// Язык распознается по всему слову, за вычетом двух символов.
		// Два символа попадают в удаления, среди тестируемых огрызков будут те,
		// что содержат только правильные символы.
		// Таким образом, ошибки с опечатками до двух символов, набранных в неправильной раскладке,
		// могут быть исправлены.
		// Дальнейшая доработка - словарь замен для компенсации неправильной раскладки.
		require.Equal(t, ruLangCode, c.LangByWord("военный"))
		require.Equal(t, ruLangCode, c.LangByWord("вfенк1"))
		require.Equal(t, ruLangCode, c.LangByWord("thпру"))

		require.Equal(t, enLangCode, c.LangByWord("motorola"))
		require.Equal(t, enLangCode, c.LangByWord("1motoяrola"))
	})

	t.Run("LanguageNotDetected", func(t *testing.T) {
		// Если в слове больше 2 некорректных символов, язык - Unknown.
		require.Equal(t, unknownLangCode, c.LangByWord("вfенк12"))
		require.Equal(t, unknownLangCode, c.LangByWord("1motoя2rola"))

		// Если в слове не большинство корректных символов, язык - Unknown.
		require.Equal(t, unknownLangCode, c.LangByWord("вfф1"))
		require.Equal(t, unknownLangCode, c.LangByWord("thпр"))

		require.Equal(t, unknownLangCode, c.LangByWord("the phrase of words"))
		require.Equal(t, unknownLangCode, c.LangByWord("это не одно слово"))
	})
}
