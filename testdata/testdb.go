package testdata

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	contpgx "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/cannonflesh/wordspell/internal/postgres"
)

type PgConn struct {
	Conn      *postgres.Conn
	Terminate func()
}

func NewPgConn(t *testing.T) *PgConn {
	ctx := context.Background()
	err := os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	require.NoError(t, err)

	dbName := "pgdb"
	dbUser := "user"
	dbPassword := "pass"

	pgC, err := contpgx.RunContainer(ctx,
		testcontainers.WithImage("postgres:16.8-bookworm"),
		contpgx.WithInitScripts(filepath.Join(ThisDir(), "basedump.sql")),
		contpgx.WithDatabase(dbName),
		contpgx.WithUsername(dbUser),
		contpgx.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(time.Minute)),
	)
	require.NoError(t, err)

	connStr, err := pgC.ConnectionString(ctx)
	require.NoError(t, err)

	re := regexp.MustCompile("postgres://([^/]+)/")
	matches := re.FindStringSubmatch(connStr)
	hostport := strings.Split(matches[1], "@")
	hostport = strings.Split(hostport[len(hostport)-1], ":")
	host := hostport[0]
	port := 5432
	if len(hostport) > 1 {
		port, err = strconv.Atoi(hostport[1])
		require.NoError(t, err)
	}

	opt := &postgres.Options{
		Host:   host,
		Port:   port,
		User:   dbUser,
		Pass:   dbPassword,
		DBName: dbName,
	}

	db, err := postgres.New(opt)
	require.NoError(t, err)
	require.NotNil(t, db)

	return &PgConn{
		Conn: db,
		Terminate: func() {
			if err := pgC.Terminate(ctx); err != nil {
				t.Fail()
			}
		},
	}
}
