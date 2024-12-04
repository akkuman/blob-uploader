/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/akkuman/blob-uploader/oci"
	"github.com/akkuman/blob-uploader/pkg/regctl"
	"github.com/akkuman/blob-uploader/pkg/util"
	"github.com/akkuman/blob-uploader/storage"
	"github.com/spf13/cobra"
)

type UploadCommandOpt struct {
	tgzFilePath string
	refName string
	username string
	password string
}

var uploadCommandOpt UploadCommandOpt

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload file to github packages",
	Long: `upload file bolb to github packages

ref: https://github.com/Homebrew/brew/blob/b753315b0b1e78b361612bf4985502bf9dca5582/Library/Homebrew/github_packages.rb#L196-L428`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !strings.HasPrefix(uploadCommandOpt.refName, "ghcr.io") {
			return fmt.Errorf("ref-name must start with ghcr.io")
		}
		if uploadCommandOpt.tgzFilePath == ""|| !util.FileExist(uploadCommandOpt.tgzFilePath) {
			return fmt.Errorf("%s is not exist", uploadCommandOpt.tgzFilePath)
		}
		reg := regctl.NewRegistry("ghcr.io", uploadCommandOpt.username, uploadCommandOpt.password)
		err := reg.Login()
		if err != nil {
			return fmt.Errorf("connect github package error: %v", err)
		}
		ociInstance := oci.NewOCI()
		stge := storage.NewGithubPackageStorage(ociInstance, reg)
		f, err := os.Open(uploadCommandOpt.tgzFilePath)
		if err != nil {
			return err
		}
		defer f.Close()
		err = stge.Upload(context.Background(), uploadCommandOpt.refName, f)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)

	uploadCmd.Flags().StringVarP(&uploadCommandOpt.tgzFilePath, "tgz-file", "f", "", "file path for tgz which will be uploaded")
	uploadCmd.Flags().StringVarP(&uploadCommandOpt.refName, "ref-name", "r", "", "the ref that you will push, exmaple: ghcr.io/example/hello:1.2.0")
	uploadCmd.Flags().StringVarP(&uploadCommandOpt.username, "username", "u", "", "the username of registry")
	uploadCmd.Flags().StringVarP(&uploadCommandOpt.password, "passowrd", "p", "", "the passowrd of registry")
}
