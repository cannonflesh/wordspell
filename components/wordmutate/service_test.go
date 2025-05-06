package wordmutate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_Deletes(t *testing.T) {
	s := New()

	dels := s.Deletes("преображение")
	require.Len(t, dels, 145)
	require.Contains(t, dels, "реображение")
	require.Contains(t, dels, "реображени")
	require.Contains(t, dels, "преобраени")
	require.Contains(t, dels, "преображени")
	require.Contains(t, dels, "преображение")
	require.NotContains(t, dels, "преобжени")
}

func TestService_InsertsEn(t *testing.T) {
	s := New()

	addOne := s.InsertRuneEn("internaionaization")
	require.Len(t, addOne, 551)
	require.Contains(t, addOne, "internationaization")
	require.Contains(t, addOne, "intern-aionaization")
	require.Contains(t, addOne, "internaionaiza-tion")
	require.NotContains(t, addOne, "internaonaiza-tion")

	var addTwo []string
	for _, w := range addOne {
		addTwoChunk := s.InsertRuneEn(w)
		addTwo = append(addTwo, addTwoChunk...)
	}

	allAdds := append(addOne, addTwo...)
	require.Len(t, allAdds, 320131)

	require.Contains(t, allAdds, "internationalization")
	require.Contains(t, allAdds, "intern-aionaiza-tion")
	require.Contains(t, allAdds, "internaionaiza-tion")
	require.NotContains(t, allAdds, "internaonaiza-tion")
}

func TestService_InsertsRu(t *testing.T) {
	s := New()

	t.Run("17LettersWord", func(t *testing.T) {
		word := "интенационализаия"
		addOne := s.InsertRuneRu(word)
		require.Len(t, addOne, 612)
		require.Contains(t, addOne, "интенационализация")
		require.Contains(t, addOne, "интернационализаия")
		require.Contains(t, addOne, "интенационали-заия")
		require.Contains(t, addOne, "интенационали-заия")
		require.Contains(t, addOne, "интенационали-заия")
		require.NotContains(t, addOne, "интеационали-заия")

		var addTwo []string
		for _, w := range addOne {
			addTwoChunk := s.InsertRuneRu(w)
			addTwo = append(addTwo, addTwoChunk...)
		}

		allAdds := append(addOne, addTwo...)
		require.Len(t, allAdds, 395964)

		require.Contains(t, allAdds, "интернационализация")
		require.Contains(t, allAdds, "инте-национали-заия")
		require.Contains(t, allAdds, "интенационали-заия-")
		require.Contains(t, allAdds, "-интенационали-заия")
		require.NotContains(t, allAdds, "-интеационали-заия")
	})

	t.Run("5LettersWord", func(t *testing.T) {
		word := "игршк"
		addOne := s.InsertRuneRu(word)
		require.Len(t, addOne, 204)
		require.Contains(t, addOne, "игршки")
		require.Contains(t, addOne, "тигршк")
		require.NotContains(t, addOne, "тигрушк")

		var addTwo []string
		for _, w := range addOne {
			addTwoChunk := s.InsertRuneRu(w)
			addTwo = append(addTwo, addTwoChunk...)
		}

		allAdds := append(addOne, addTwo...)
		require.Len(t, allAdds, 48756)

		require.Contains(t, allAdds, "игрушки")
		require.Contains(t, allAdds, "игрушка")
		require.Contains(t, allAdds, "играшка")
		require.NotContains(t, allAdds, "играчка")
	})

	t.Run("24LettersWord", func(t *testing.T) {
		word := "интернационализация-каюк"
		addOne := s.InsertRuneRu(word)
		require.Len(t, addOne, 850)

		var addTwo []string
		for _, w := range addOne {
			addTwoChunk := s.InsertRuneRu(w)
			addTwo = append(addTwo, addTwoChunk...)
		}

		allAdds := append(addOne, addTwo...)
		require.Len(t, allAdds, 752250)
	})
}
