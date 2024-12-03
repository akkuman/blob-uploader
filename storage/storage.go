package storage

import (
	"context"
	"io"
)

type Metadata struct {
	Arch string
	OS string
}

type Storage interface {
	Upload(ctx context.Context, uri string, reader io.Reader) error
	Download(ctx context.Context, uri string, writer io.Writer) error
}
