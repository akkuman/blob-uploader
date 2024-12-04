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
	"github.com/tidwall/gjson"
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

func (s *GithubPackageStorage) Upload(ctx context.Context, imageRef string, platform util.Platform, imageSource string, reader io.Reader) error {
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
	err = s.ociInstance.BuildOCI(ctx, platform, blobFilePath, s.registry.GetVersion(imageRef), imageSource)
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

func (s *GithubPackageStorage) Download(ctx context.Context, imageRef string, platform util.Platform, writer io.Writer) error {
	r, err := ref.New(imageRef)
	if err != nil {
		return err
	}
	rg := regctl.NewAnonymousRegistry()
	if r.Tag == "latest" {
		tags, err := rg.GetTags(ctx, imageRef)
		if err != nil {
			return err
		}
		r.Tag = tags[len(tags)-1]
	}
	refName := fmt.Sprintf("%s/%s:%s", r.Registry, r.Repository, r.Tag)
	manifest, err := rg.GetManifest(ctx, refName)
	if err != nil {
		return err
	}
	var fileDigest string
	for _, mf := range gjson.Get(manifest, "manifests").Array() {
		if mf.Get("platform.architecture").String() == platform.Arch && mf.Get("platform.os").String() == platform.OS {
			fileDigest = mf.Get(`annotations.dev\.pkgforge\.bin\.digest`).String()
			break
		}
	}
	return rg.DownloadBlob(ctx, refName, fileDigest, writer)
}
