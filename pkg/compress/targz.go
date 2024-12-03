package compress

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
)

func CompressToTmpFile(files []string) (targzPath string, err error) {
	out, err := os.CreateTemp("", "blob-uploader.*.tar.gz")
	if err != nil {
		return "", err
	}
	defer out.Close()
	err = Compress(files, out)
	return out.Name(), err
}

func Compress(files []string, out io.Writer) error {
	gw := gzip.NewWriter(out)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, file := range files {
		err := addToArchive(tw, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func addToArchive(tw *tar.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	header.Name = filename

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}