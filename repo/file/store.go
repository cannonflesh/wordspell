package file

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// Store - file-based хранилище данных.
type Store struct {
	dir string
}

// New конструктор file-based хранилища данных.
func New(opt *Options) *Store {
	return &Store{
		dir: opt.DataDir,
	}
}

///// Имплементация интерфейса index.dataStore /////

func (s *Store) IsExist(key string) (bool, error) {
	fPath := filepath.Join(s.dir, key)

	res, err := os.Stat(fPath)
	if err == nil {
		if !res.Mode().IsRegular() {
			return false, errors.New("this is not regular file")
		}

		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, errors.WithStack(err)
}

func (s *Store) DataReader(key string) (io.ReadCloser, error) {
	fPath := filepath.Join(s.dir, key)

	if exists, err := s.IsExist(key); !exists {
		return nil, err
	}

	fh, err := os.Open(fPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return fh, nil
}

func (s *Store) Save(key string, content io.Reader) error {
	wh, err := s.indexFileWriteHandler(key)
	if err != nil {
		return err
	}
	defer func() {
		_ = wh.Close()
	}()

	_, err = io.Copy(wh, content)

	return errors.WithStack(err)
}

func (s *Store) indexFileWriteHandler(key string) (io.WriteCloser, error) {
	path := filepath.Join(s.dir, key)
	fh, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return fh, nil
}
