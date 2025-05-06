package file

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const checkDataFile = "check.index"

func TestStore_IsExists_Save_DataReader(t *testing.T) {
	opt := &Options{DataDir: "/tmp"}

	store := New(opt)

	initial := "initial content"
	actual := "actual"

	exists, err := store.IsExist(checkDataFile)
	require.NoError(t, err)
	require.False(t, exists)

	err = store.Save(checkDataFile, bytes.NewBufferString(initial))
	require.NoError(t, err)
	defer func() {
		fPath := filepath.Join(opt.DataDir, checkDataFile)
		_ = os.Remove(fPath)
	}()

	dStream, err := store.DataReader(checkDataFile)
	require.NoError(t, err)

	data, err := io.ReadAll(dStream)
	require.Equal(t, initial, string(data))
	err = dStream.Close()
	require.NoError(t, err)

	err = store.Save(checkDataFile, bytes.NewBufferString(actual))
	require.NoError(t, err)

	dStream, err = store.DataReader(checkDataFile)
	require.NoError(t, err)

	data, err = io.ReadAll(dStream)
	require.Equal(t, actual, string(data))
	err = dStream.Close()
	require.NoError(t, err)
}
