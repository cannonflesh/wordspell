package s3

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/pkg/errors"
)

// Store репозиторий s3 хранилища.
type Store struct {
	cli      *s3.Client
	uploader *manager.Uploader
	opts     Options
}

// NewStore вернет новый инстанс репозитория s3 хранилища.
func NewStore(cli *s3.Client, opts Options) (*Store, error) {
	_, err := cli.HeadBucket(context.Background(),
		&s3.HeadBucketInput{
			Bucket: aws.String(opts.Bucket),
		},
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r := &Store{
		cli:      cli,
		uploader: manager.NewUploader(cli),
		opts:     opts,
	}

	return r, nil
}

///// Имплементация интерфейса index.dataStore /////

// DataReader отдает io.ReadCloser, ответственность за закрытие - на вызывающей стороне.
func (r *Store) DataReader(key string) (io.ReadCloser, error) {
	ctx := context.Background()

	s3o, err := r.cli.GetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(r.opts.Bucket),
			Key:    aws.String(key),
		},
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return s3o.Body, nil
}

// IsExist возвращает true или false, когда файл существует или не существует соответственно.
// В случае возникновения ошибки, не связанной с существованием файла, возвращается ошибка, в противном случае nil.
func (r *Store) IsExist(key string) (bool, error) {
	_, err := r.cli.HeadObject(
		context.Background(),
		&s3.HeadObjectInput{
			Bucket: aws.String(r.opts.Bucket),
			Key:    aws.String(key),
		},
	)

	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			if _, ok := apiError.(*types.NotFound); ok {
				err = nil
			}
		}

		return false, errors.Wrap(err, "key: "+key)
	}

	return true, nil
}

// Save сохраняет файл в s3, возвращает ошибку при её возникновени, в противном случае nil.
// Ответственность за закрытие переданного ридера - на вызывающей стороне, тут это просто ридер.
func (r *Store) Save(key string, content io.Reader) error {
	_, err := r.uploader.Upload(
		context.Background(),
		&s3.PutObjectInput{
			Bucket: aws.String(r.opts.Bucket),
			Key:    aws.String(key),
			Body:   content,
		},
	)

	return errors.WithStack(err)
}
