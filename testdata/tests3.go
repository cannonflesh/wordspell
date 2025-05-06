package testdata

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/romnn/testcontainers/minio"
	"github.com/stretchr/testify/require"
)

type S3Provider struct {
	Endpoint string
	Key      string
	Secret   string

	TermFn func()
}

func NewTestS3(t *testing.T) *S3Provider {
	user := "minio"
	pass := "miniosecret"
	err := os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	require.NoError(t, err)

	ctx := context.Background()
	opts := minio.Options{
		RootUser:     user,
		RootPassword: pass,
	}
	opts.ContainerOptions.StartupTimeout = time.Minute
	opts.Image = "minio/minio:latest"
	container, err := minio.Start(ctx, opts)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	return &S3Provider{
		Endpoint: container.ConnectionURI(),
		Key:      user,
		Secret:   pass,

		TermFn: func() { container.Terminate(ctx) },
	}
}
