package catalog

import (
	"github.com/cannonflesh/wordspell/testdata"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCatalog_TradeMarkNames(t *testing.T) {
	conn := testdata.NewPgConn(t)
	defer conn.Terminate()

	lgr := logrus.NewEntry(logrus.New())
	repo := New(conn.Conn, lgr)

	t.Run("TradeMarkNames", func(t *testing.T) {
		names, err := repo.TradeMarkNames(10, 3)
		require.NoError(t, err)

		require.Equal(t, []string{"Lucas Film", "MINAKU", "Nokian"}, names)
	})

	t.Run("CategoryNames", func(t *testing.T) {
		names, err := repo.CategoryNames(10, 3)
		require.NoError(t, err)

		require.Equal(t, []string{"Шкатулки3", "Игрушки", "Велосипеды"}, names)
	})

	t.Run("ItemData", func(t *testing.T) {
		name, desc, err := repo.ItemData(20, 1)
		require.NoError(t, err)
		require.Len(t, name, 1)
		require.Len(t, desc, 1)
		require.Equal(t, "very very long name what will take more than one line", name[0])
		require.Contains(t, desc[0], "<p>По теории фэн-шуй, для устранения домашних проблем нужно настроить правильную тональность дома.")
	})
}
