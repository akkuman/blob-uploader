/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/akkuman/blob-uploader/pkg/util"
	"github.com/akkuman/blob-uploader/storage"
	"github.com/regclient/regclient/types/ref"
	"github.com/spf13/cobra"
)

type DownloadCommandOpt struct {
	outFile string
	refName string
	platform string
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
		if _, err := ref.New(downloadCommandOpt.refName); err != nil {
			return err
		}
		
		platform := util.ParsePlatform(downloadCommandOpt.platform)
		if platform == nil {
			return fmt.Errorf("%s is not allowed", downloadCommandOpt.platform)
		}
		stge := storage.NewGithubPackageStorage(nil, nil)
		w, err := os.Create(downloadCommandOpt.outFile)
		if err != nil {
			return err
		}
		defer w.Close()
		err = stge.Download(context.Background(), downloadCommandOpt.refName, *platform, w)
		if err != nil {
			return err
		}
		fmt.Println("Successfully download tgz from registry!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().StringVarP(&downloadCommandOpt.outFile, "out-file", "o", "", "file path for tgz")
	downloadCmd.Flags().StringVarP(&downloadCommandOpt.refName, "ref-name", "r", "", "the ref that you want download from (e.g.: ghcr.io/example/hello:1.2.0)")
	downloadCmd.Flags().StringVarP(&downloadCommandOpt.platform, "platform", "", "linux/amd64", "Specify platform (e.g. linux/amd64)")

	requires := []string{
		"out-file",
		"ref-name",
	}

	for _, i := range requires {
		downloadCmd.MarkFlagRequired(i)
	}
}
