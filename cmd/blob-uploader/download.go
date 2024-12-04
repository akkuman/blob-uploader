/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/akkuman/blob-uploader/pkg/regctl"
	"github.com/regclient/regclient/types/ref"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type DownloadCommandOpt struct {
	outFile string
	refName string
}

var downloadCommandOpt DownloadCommandOpt

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download file from github packages",
	Long: `download file bolb from github packages,
	
ref: https://github.com/orgs/Homebrew/discussions/4335
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		downloadCommandOpt.refName = strings.ToLower(downloadCommandOpt.refName)
		if !strings.HasPrefix(downloadCommandOpt.refName, "ghcr.io") {
			return fmt.Errorf("ref-name must start with ghcr.io")
		}
		r, err := ref.New(downloadCommandOpt.refName)
		if err != nil {
			return err
		}
		ctx := context.Background()
		rg := regctl.NewAnonymousRegistry()
		if r.Tag == "latest" {
			tags, err := rg.GetTags(ctx, downloadCommandOpt.refName)
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
			if mf.Get("platform.architecture").String() == runtime.GOARCH && mf.Get("platform.os").String() == runtime.GOOS {
				fileDigest = mf.Get(`annotations.dev\.pkgforge\.bin\.digest`).String()
				break
			}
		}
		w, err := os.Create(downloadCommandOpt.outFile)
		if err != nil {
			return err
		}
		defer w.Close()
		return rg.DownloadBlob(ctx, refName, fileDigest, w)
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().StringVarP(&downloadCommandOpt.outFile, "out-file", "o", "", "file path for tgz")
	downloadCmd.Flags().StringVarP(&downloadCommandOpt.refName, "ref-name", "r", "", "the ref that you want download from, example: ghcr.io/example/hello:1.2.0")
}
