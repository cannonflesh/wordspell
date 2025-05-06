package postgres

import (
	"context"
	"net/url"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

type Conn struct {
	*pgx.Conn
}

func New(opt *Options) (*Conn, error) {
	conn, err := pgx.Connect(context.Background(), dsn(opt))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	res := &Conn{
		Conn: conn,
	}

	return res, nil
}

// dsn строим postgres dsn с помощью net/url.
func dsn(opt *Options) string {
	var (
		userInfo *url.Userinfo
	)

	if opt.User != "" && opt.Pass != "" {
		userInfo = url.UserPassword(opt.User, opt.Pass)
	} else if opt.User != "" {
		userInfo = url.User(opt.User)
	}

	res := &url.URL{
		Scheme: "postgres",
		User:   userInfo,
		Host:   opt.Host + ":" + strconv.Itoa(opt.Port),
		Path:   opt.DBName,
	}

	return res.String()
}
