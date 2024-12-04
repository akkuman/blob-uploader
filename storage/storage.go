package storage

import (
	"context"
	"io"

	"github.com/akkuman/blob-uploader/pkg/util"
)

type Metadata struct {
	Arch string
	OS string
}

type Storage interface {
	Upload(ctx context.Context, imageRef string, platform util.Platform, reader io.Reader) error
	Download(ctx context.Context, imageRef string, platform util.Platform, writer io.Writer) error
}
