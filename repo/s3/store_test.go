package s3

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"

	s3client "gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/internal/s3"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/testdata"
)

func TestStore_IsExists_DataReader(t *testing.T) {
	testClient := testdata.NewTestS3(t)

	defer testClient.TermFn()

	opts := Options{
		Bucket: "test-bucket",
	}

	content := bytes.NewBufferString("first-write-try")
	key := "ru.index"

	cli, err := s3client.NewClient(s3client.Options{
		Endpoint:        testClient.Endpoint,
		AccessKeyID:     testClient.Key,
		SecretAccessKey: testClient.Secret,
	})
	require.NoError(t, err)

	_, err = cli.CreateBucket(
		context.Background(),
		&s3.CreateBucketInput{
			Bucket: aws.String(opts.Bucket),
		},
	)
	require.NoError(t, err)

	repo, err := NewStore(cli, opts)
	require.NoError(t, err)

	// поначалу медиаданных в бакете нет.
	ok, err := repo.IsExist(key)
	require.NoError(t, err)
	require.False(t, ok)

	// запишем в бакет данные.
	err = repo.Save(key, content)
	require.NoError(t, err)

	// перепишем их еще раз, поверх существующей записи.
	// Ошибки нет, запись производится успешно.
	content = bytes.NewBufferString("indexdata")
	err = repo.Save(key, content)
	require.NoError(t, err)

	// теперь в бакете есть данные.
	ok, err = repo.IsExist(key)
	require.NoError(t, err)
	require.True(t, ok)

	// попробуем извлечь.
	res, err := repo.DataReader(key)
	require.NoError(t, err)
	testBin, err := io.ReadAll(res)
	require.NoError(t, err)
	require.Equal(t, "indexdata", string(testBin))
	require.NoError(t, res.Close())
}
