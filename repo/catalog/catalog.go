package catalog

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/domain"
)

type pgxDriver interface {
	Query(ctx context.Context, queue string, args ...any) (pgx.Rows, error)
}

type Repository struct {
	conn pgxDriver
	log  *logrus.Entry
}

func New(conn pgxDriver, l *logrus.Entry) *Repository {
	return &Repository{
		conn: conn,
		log:  l.WithField(domain.CategoryFieldName, "repo.site-catalog"),
	}
}

func (c *Repository) TradeMarkNames(start, limit int) ([]string, error) {
	var res []string

	ctx := context.Background()
	rows, err := c.conn.Query(ctx, readTrademarkNamesSQL, start, limit)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.WithStack(err)
	}

	chunk, err := parseNameRows(rows)
	if err != nil {
		return nil, err
	}

	res = append(res, chunk...)

	return res, nil
}

const readTrademarkNamesSQL = `SELECT name
FROM trademark
ORDER BY name
OFFSET $1
LIMIT $2`

func (c *Repository) CategoryNames(start, limit int) ([]string, error) {
	var res []string

	ctx := context.Background()
	rows, err := c.conn.Query(ctx, readCategoryNamesSQL, start, limit)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.WithStack(err)
	}

	chunk, err := parseNameRows(rows)
	if err != nil {
		return nil, err
	}

	res = append(res, chunk...)

	return res, nil
}

const readCategoryNamesSQL = `SELECT name
FROM category
ORDER BY id
OFFSET $1
LIMIT $2`

func (c *Repository) ItemData(start, limit int) ([]string, []string, error) {
	ctx := context.Background()

	rows, err := c.conn.Query(ctx, readItemDataSQL, start, limit)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, errors.WithStack(err)
	}

	return parseDataRows(rows)
}

const readItemDataSQL = `SELECT name, coalesce(description, '')
FROM item
ORDER BY id
OFFSET $1
LIMIT $2`

func parseNameRows(rows pgx.Rows) ([]string, error) {
	var (
		res []string
		err error
	)

	defer func() {
		rows.Close()
		if rowsErr := rows.Err(); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			err = errors.WithStack(rowsErr)
		}
	}()

	for rows.Next() {
		var n string
		err := rows.Scan(&n)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		res = append(res, n)
	}

	return res, err
}

func parseDataRows(rows pgx.Rows) ([]string, []string, error) {
	var (
		name, desc []string
		err        error
	)

	defer func() {
		rows.Close()
		if rowsErr := rows.Err(); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			err = errors.WithStack(rowsErr)
		}
	}()

	for rows.Next() {
		var (
			n, d string
		)
		err := rows.Scan(&n, &d)
		if err != nil {
			return nil, nil, errors.WithStack(err)
		}

		name = append(name, n)
		if d != "" {
			desc = append(desc, d)
		}
	}

	return name, desc, err
}
