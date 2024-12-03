package oci

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/akkuman/blob-uploader/pkg/util"
)

type OCI struct {
	rootDir string
	blobsDir string
}

func NewOCI() *OCI {
	ociRootDir, err := os.MkdirTemp("", "oci")
	if err != nil {
		panic(err)
	}
	blobsDir := filepath.Join(ociRootDir, "blobs/sha256")
	err = os.MkdirAll(blobsDir, 0775)
	if err != nil {
		panic(err)
	}
	return &OCI{
		rootDir: ociRootDir,
		blobsDir: blobsDir,
	}
}

func (s *OCI) GetRootDir() string {
	return s.rootDir
}

func (s *OCI) Close() error {
	return os.RemoveAll(s.rootDir)
}

func (s *OCI) writeMap(ctx context.Context, dir string, data map[string]any, filename string) (newFName string, fileLen int, err error) {
	var jsonBytes []byte
	jsonBytes, err = json.Marshal(data)
	if err != nil {
		return
	}
	if filename == "" {
		filename, err = util.GetSHA256(bytes.NewReader(jsonBytes))
		if err != nil {
			return
		}
	}
	err = os.WriteFile(filepath.Join(dir, filename), jsonBytes, 0664)
	return filename, len(jsonBytes), err
}

func (s *OCI) writeImageLayout(ctx context.Context) error {
	data := map[string]any {
		"imageLayoutVersion": "1.0.0",
	}
	_, _, err := s.writeMap(ctx, s.rootDir, data, "oci-layout")
	return err
}

func (s *OCI) writeBlobs(ctx context.Context, blobFilepath string) (hexdigest string, err error) {
	hexdigest, err = util.CalcFileSHA256(blobFilepath)
	if err != nil {
		return
	}
	err = util.CopyFile(blobFilepath, filepath.Join(s.blobsDir, hexdigest))
	return
}

func (s *OCI) writeImageConfig(ctx context.Context, baseMap map[string]any, tarSHA256 string) (jsonSHA256 string, jsonSize int, err error) {
	dstMap := map[string]any{
		"rootfs": map[string]any{
			"type": "layers",
			"diff_ids": []string{
				fmt.Sprintf("sha256:%s", tarSHA256),
			},
		},
	}
	maps.Copy(dstMap, baseMap)
	return s.writeMap(ctx, s.blobsDir, dstMap, "")
}

func (s *OCI) writeImageIndex(ctx context.Context, manifests []map[string]any, annotations map[string]any) (jsonSHA256 string, jsonSize int, err error) {
	imageIndex := map[string]any{
		"schemaVersion": 2,
		"manifests":     manifests,
		"annotations":   annotations,
	}
	return s.writeMap(ctx, s.blobsDir, imageIndex, "")
}

func (s *OCI) writeIndexJSON(ctx context.Context, indexJSONSHA256 string, indexJSONSize int, annotations map[string]any) error {
	indexJSON := map[string]any{
		"schemaVersion": 2,
		"manifests": []map[string]any{{
			"mediaType":   "application/vnd.oci.image.index.v1+json",
			"digest":      fmt.Sprintf("sha256:%s", indexJSONSHA256),
			"size":        indexJSONSize,
			"annotations": annotations,
		}},
	}
	_, _, err := s.writeMap(ctx, s.rootDir, indexJSON, "index.json")
	return err
}

func (s *OCI) BuildOCI(ctx context.Context, targzFilePath string, tagVersion string) (err error) {
	err = s.writeImageLayout(ctx)
	if err != nil {
		return
	}
	var targzSHA256 string
	targzSHA256, err = s.writeBlobs(ctx, targzFilePath)
	if err != nil {
		return
	}
	platformMap := map[string]any{
		"architecture": "amd64",
		"os":           "linux",
	}
	var tarSHA256 string
	tarSHA256, err = util.GetTarSHA256FromGz(targzFilePath)
	if err != nil {
		return
	}
	var jsonSHA256 string
	var jsonSize int
	jsonSHA256, jsonSize, err = s.writeImageConfig(ctx, platformMap, tarSHA256)
	if err != nil {
		return
	}
	blobFileSize, err := util.GetFileSize(targzFilePath)
	if err != nil {
		return
	}
	annotations := map[string]any{
		"dev.pkgforge.bin.digest": targzSHA256,
	}
	imageManifest := map[string]any{
		"schemaVersion": 2,
		"config": map[string]any{
			"mediaType": "application/vnd.oci.image.config.v1+json",
			"digest":    fmt.Sprintf("sha256:%s", jsonSHA256),
			"size":      jsonSize,
		},
		"layers": []any{
			map[string]any{
				"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
				"digest":    fmt.Sprintf("sha256:%s", targzSHA256),
				"size":      blobFileSize,
			},
		},
		"annotations": annotations,
	}
	manifestJSONSHA256, manifestJSONSize, err := s.writeMap(ctx, s.blobsDir, imageManifest, "")
	if err != nil {
		return err
	}
	manifest := map[string]any{
		"mediaType":   "application/vnd.oci.image.manifest.v1+json",
		"digest":      fmt.Sprintf("sha256:%s", manifestJSONSHA256),
		"size":        manifestJSONSize,
		"platform":    platformMap,
		"annotations": annotations,
	}
	indexJSONSHA256, indexJSONSize, err := s.writeImageIndex(ctx, []map[string]any{manifest}, annotations)
	if err != nil {
		return err
	}
	err = s.writeIndexJSON(ctx, indexJSONSHA256, indexJSONSize, map[string]any{
		"org.opencontainers.image.ref.name": "latest",
	})
	return err
}
