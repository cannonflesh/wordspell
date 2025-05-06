package main

import (
	"context"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/testdata"
	"sync"
	"time"

	"github.com/cannonflesh/microprof"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/components/bloomfilter"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/internal/postgres"
	s3client "gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/internal/s3"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/options"
	s3source "gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/repo/s3"
)

const batchSize = 1000

func main() {
	lgr := logrus.NewEntry(logrus.New())

	opt := &options.Options{
		Bloom: bloomfilter.Options{FalsePositiveRate: 0.01},
		SiteDB: postgres.Options{
			Host:   "postgres",
			Port:   5432,
			DBName: "pgdb",
			User:   "user",
			Pass:   "pass",
		},
		S3Client: s3client.Options{
			Endpoint:        "minio:9000",
			AccessKeyID:     "minio",
			SecretAccessKey: "miniosecret",
		},
		S3Data: s3source.Options{
			Bucket: "wordspell-index",
			Name:   "wordspell",
		},
	}

	ws, err := wordspell.New(opt, lgr)
	if err != nil {
		lgr.Fatal(err)
	}

	conn, err := postgres.New(&opt.SiteDB)
	if err != nil {
		lgr.Fatal(err)
	}

	wg := new(sync.WaitGroup)

	total := 0
	startCorrect := time.Now()
	startBatchCorrect := time.Now()
	uniq := make(map[string]bool)

	reqSrc, err := testdata.SearchRequests()
	if err != nil {
		lgr.Fatal(err)
	}

	var maxCorrectTime time.Duration
	corrected := make(map[string]string, batchSize)
	for sr := range reqSrc {
		if uniq[sr] {
			continue
		}

		uniq[sr] = true

		if len(corrected) == batchSize {
			wg.Add(1)
			go func(batch map[string]string) {
				defer wg.Done()
				err := saveBatch(conn, batch)
				if err != nil {
					lgr.WithError(err).Error("saving batch")
				}
			}(corrected)
			corrected = make(map[string]string, batchSize)

			lgr.Infof("corrected %d requests in %v, total: %d", batchSize, time.Since(startBatchCorrect), total)
			startBatchCorrect = time.Now()
		}

		startReqCorrect := time.Now()
		corrected[sr] = ws.Correct(sr)
		reqCorrectElapsed := time.Since(startReqCorrect)
		if reqCorrectElapsed > maxCorrectTime {
			maxCorrectTime = reqCorrectElapsed
		}

		total++

	}

	microprof.PrintProfilingInfo(lgr, microprof.UnitsMb, false)

	wg.Wait()

	if len(corrected) > 0 {
		err = saveBatch(conn, corrected)
		if err != nil {
			lgr.Fatal(err)
		}
	}

	lgr.Infof("corrected total of %d requests in %v, max correction time per request: %v", total, time.Since(startCorrect), maxCorrectTime)
}

const saveCorrectionsSQL = "INSERT INTO search_req_correct (src_req, corrected) VALUES ($1, $2)"

func saveBatch(conn *postgres.Conn, data map[string]string) error {
	ctx := context.Background()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	for m, c := range data {
		_, err = tx.Exec(ctx, saveCorrectionsSQL, m, c)
		if err != nil {
			rbackErr := tx.Rollback(ctx)
			if rbackErr != nil {
				return errors.Wrap(err, rbackErr.Error())
			}

			return errors.WithStack(err)
		}
	}

	return errors.WithStack(tx.Commit(ctx))
}
