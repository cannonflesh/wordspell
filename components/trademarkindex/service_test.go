package trademarkindex

import (
	"bytes"
	"io"
	"testing"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestComponent_New(t *testing.T) {
	lgr := logrus.NewEntry(logrus.New())

	t.Run("Success", func(t *testing.T) {
		store := NewMockDataStore(t)
		store.EXPECT().DataReader(storeKey).
			Return(goldenIndex(), nil).
			Once()
		store.EXPECT().IsExist(storeKey).
			Return(true, nil).
			Once()

		comp, err := NewService(store, lgr)
		require.NoError(t, err)
		require.NotNil(t, comp)
	})
	t.Run("NoTrademarkIndexFound", func(t *testing.T) {
		store := NewMockDataStore(t)
		store.EXPECT().IsExist(storeKey).
			Return(false, nil).
			Once()

		comp, err := NewService(store, lgr)
		require.EqualError(t, err, "no trademark index found")
		require.Nil(t, comp)
	})
	t.Run("ErrorReadingTrademarkIndex", func(t *testing.T) {
		foundErr := errors.New("index-found-err")
		store := NewMockDataStore(t)
		store.EXPECT().IsExist(storeKey).
			Return(false, foundErr).
			Once()

		comp, err := NewService(store, lgr)
		require.ErrorIs(t, err, foundErr)
		require.Nil(t, comp)
	})
}

func TestComponent_Find(t *testing.T) {
	lgr := logrus.NewEntry(logrus.New())

	store := NewMockDataStore(t)
	store.EXPECT().DataReader(storeKey).
		Return(goldenIndex(), nil).
		Once()
	store.EXPECT().IsExist(storeKey).
		Return(true, nil).
		Once()

	comp, err := NewService(store, lgr)
	require.NoError(t, err)
	require.NotNil(t, comp)

	t.Run("NotFound", func(t *testing.T) {
		hstack := []string{"head", "Mazda", "super", "puper", "tail", "tail"}
		found, tail := comp.Find(hstack)
		require.Empty(t, found)
		require.Equal(t, hstack, tail)
	})

	t.Run("FoundShortPresent", func(t *testing.T) {
		hstack := []string{"Mazda", "super", "tail", "tail"}
		found, tail := comp.Find(hstack)
		require.Equal(t, "@Mazda#super", found)
		require.Equal(t, []string{"tail", "tail"}, tail)
	})

	t.Run("FoundLongest", func(t *testing.T) {
		hstack := []string{"Mazda", "super", "puper", "tail", "tail"}
		found, tail := comp.Find(hstack)
		require.Equal(t, "@Mazda#super#puper", found)
		require.Equal(t, []string{"tail", "tail"}, tail)
	})

	t.Run("FoundHeadOnly", func(t *testing.T) {
		hstack := []string{"Mazda", "tail", "tail"}
		found, tail := comp.Find(hstack)
		require.Equal(t, "@Mazda", found)
		require.Equal(t, []string{"tail", "tail"}, tail)
	})

	t.Run("NoOneWordAllowed", func(t *testing.T) {
		hstack := []string{"Cooper", "tail", "tail"}
		found, tail := comp.Find(hstack)
		require.Empty(t, found)
		require.Equal(t, hstack, tail)
	})

	t.Run("PartialNameNotFound", func(t *testing.T) {
		hstack := []string{"Cooper", "super", "tail", "tail"}
		found, tail := comp.Find(hstack)
		require.Empty(t, found)
		require.Equal(t, hstack, tail)
	})
}

func goldenIndex() io.ReadCloser {
	data := []byte(`Mazda
Mazda super
Mazda super puper
Mazda puper duper cooper
Cooper super dooper
`,
	)

	return io.NopCloser(bytes.NewBuffer(data))
}
