package storage

import (
	"context"
	"os"
	"testing"

	"github.com/akkuman/blob-uploader/oci"
	"github.com/akkuman/blob-uploader/pkg/compress"
	"github.com/akkuman/blob-uploader/pkg/regctl"
	"github.com/akkuman/blob-uploader/pkg/util"
	_ "github.com/akkuman/blob-uploader/testinit"
)

func TestGithubPackageStorage(t *testing.T) {
	username, ok := os.LookupEnv("GITHUB_USER")
	if !ok {
		t.Error("GITHUB_USER environment must to be set")
		return
	}
	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		t.Error("GITHUB_TOKEN environment must to be set")
		return
	}
	reg := regctl.NewRegistry("ghcr.io", username, token)
	err := reg.Login()
	if err != nil {
		t.Error("connect github package error:", err)
		return
	}
	ociInstance := oci.NewOCI()
	defer ociInstance.Close()
	s := NewGithubPackageStorage(ociInstance, reg)
	targzPath, err := compress.CompressToTmpFile([]string{"./_testdata/wget"})
	if err != nil {
		t.Error("compress to tar.gz failed:", err)
		return
	}
	defer os.Remove(targzPath)
	f, err := os.Open(targzPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()
	imageSource := "https://github.com/akkuman/blob-uploader"
	err = s.Upload(context.Background(), "akkuman/wgettest:0.0.1", util.DefaultPlatform, imageSource, f)
	if err != nil {
		t.Error("upload to github packages failed:", err)
		return
	}
}