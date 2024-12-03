package storage

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/akkuman/blob-uploader/oci"
	"github.com/akkuman/blob-uploader/pkg/regctl"
	"github.com/akkuman/blob-uploader/pkg/util"
	"github.com/regclient/regclient/types/ref"
)

type GithubPackageStorage struct {
	registry    *regctl.Registry
	ociInstance *oci.OCI
}

var _ Storage = &GithubPackageStorage{}

func NewGithubPackageStorage(ociInstance *oci.OCI, registry *regctl.Registry) *GithubPackageStorage {
	return &GithubPackageStorage{
		ociInstance: ociInstance,
		registry:    registry,
	}
}

func (s *GithubPackageStorage) Upload(ctx context.Context, imageRef string, reader io.Reader) error {
	r, err := ref.New(imageRef)
	if err != nil {
		return err
	}
	imageRef = fmt.Sprintf("%s:%s", r.Repository, r.Tag)
	blobFilePath, err := util.WriteToTempFile(reader, "blob.*.tar.gz")
	if err != nil {
		return fmt.Errorf("write blob to file failed: %w", err)
	}
	defer os.Remove(blobFilePath)
	err = s.ociInstance.BuildOCI(ctx, blobFilePath, s.registry.GetVersion(imageRef))
	if err != nil {
		return fmt.Errorf("build oci failed: %w", err)
	}
	defer s.ociInstance.Close()
	err = s.registry.ImageCopy(ctx, s.ociInstance.GetRootDir(), imageRef)
	if err != nil {
		return fmt.Errorf("image copy: %w", err)
	}
	return nil
}

func (s *GithubPackageStorage) Download(ctx context.Context, uri string, writer io.Writer) error {
	return nil
}
