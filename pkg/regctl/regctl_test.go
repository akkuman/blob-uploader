package regctl

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/akkuman/blob-uploader/pkg/util"
	"github.com/regclient/regclient/types/ref"
	"github.com/tidwall/gjson"
)

func TestLoginRegistry(t *testing.T) {
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
	rg := NewRegistry("ghcr.io", username, token)
	err := rg.Login()
	if err != nil {
		t.Error(err)
	}
}

func TestGetTags(t *testing.T) {
	for _, refName := range []string{
		"ghcr.io/homebrew/core/hello",
		"ghcr.io/homebrew/core/hello:123",
		"ghcr.io/homebrew/core/hello:2.10",
	} {
		t.Run(refName, func(t *testing.T) {
			rg := NewAnonymousRegistry()
			tags, err := rg.GetTags(context.Background(), refName)
			if err != nil {
				t.Error(err)
			}
			if len(tags) == 0 {
				t.Error("the length of tags must > 0")
			}
		})
	}
}

func TestGetTagsWithErrorRef(t *testing.T) {
	rg := NewAnonymousRegistry()
	_, err := rg.GetTags(context.Background(), "ghcr.io/homebrew/core/hello111")
	if !strings.Contains(err.Error(), "404") {
		t.Error("the status code must be 404")
	}
}

func TestGetManifest(t *testing.T) {
	rg := NewAnonymousRegistry()
	data, err := rg.GetManifest(context.Background(), "ghcr.io/homebrew/core/hello:2.10")
	if err != nil {
		t.Error(err)
	}
	if gjson.Get(data, `annotations.com\.github\.package\.type`).String() != "homebrew_bottle" {
		t.Error("error annotations")
	}
}

func TestParseRefName(t *testing.T) {
	for _, x := range []struct{
		refName string
		reg string
		repo string
		version string
	}{
		{
			"ghcr.io/homebrew/core/hello",
			"ghcr.io",
			"homebrew/core/hello",
			"latest",
		},
		{
			"ghcr.io/homebrew/core/hello:2.10",
			"ghcr.io",
			"homebrew/core/hello",
			"2.10",
		},
		{
			"wget",
			"docker.io",
			"library/wget",
			"latest",
		},
	} {
		t.Run(x.refName, func(t *testing.T) {
			r, err := ref.New(x.refName)
			if err != nil {
				t.Error(err)
			}
			if r.Registry != x.reg {
				t.Errorf("%s != %s", r.Registry, x.reg)
			}
			if r.Repository != x.repo {
				t.Errorf("%s != %s", r.Repository, x.repo)
			}
			if r.Tag != x.version {
				t.Errorf("%s != %s", r.Tag, x.version)
			}
		})
	}
}

func TestParseRefNameContainsUppercaseLetter(t *testing.T) {
	_, err := ref.New("ghcr.io/Akkuman/wget")
	if err == nil || !strings.Contains(err.Error(), "repo must be lowercase") {
		t.Error("err must contains 'repo must be lowercase'")
	}
}

func TestDownloadBlob(t *testing.T) {
	rg := NewAnonymousRegistry()
	out, err := os.CreateTemp("", "DownloadBlob.test.*")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(out.Name())
	defer out.Close()
	sha256 := "7935d0efdae69742f5140d514ef2e3e50d1d7cb82104cf6033ad51b900c12749"
	err = rg.DownloadBlob(
		context.Background(),
		"ghcr.io/homebrew/core/hello:2.12.1",
		sha256,
		out,
	)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = out.Seek(0, io.SeekStart)
	if err != nil {
		t.Error(err)
		return
	}
	hexdigest, err := util.GetSHA256(out)
	if err != nil {
		t.Error(err)
		return
	}
	if sha256 != hexdigest {
		t.Errorf("%s != %s", sha256, hexdigest)
		return
	}
}