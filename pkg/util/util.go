package util

import (
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
)

func GetSHA256(reader io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, reader); err != nil {
		return "", err
	}
	hexdigest := fmt.Sprintf("%x", h.Sum(nil))
	return hexdigest, nil
}

func CalcFileSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return GetSHA256(f)
}

func CopyFile(srcpath string, dstpath string) (err error) {
	r, err := os.Open(srcpath)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(dstpath)
	if err != nil {
		return err
	}

	defer func() {
		if c := w.Close(); c != nil {
			err = c
		}
	}()

	_, err = io.Copy(w, r)
	return err
}

func GetTarSHA256FromGz(filePath string) (string, error) {
	r, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer r.Close()
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return "", err
	}
	return GetSHA256(gzipReader)
}

func GetFileSize(filepath string) (int64, error) {
	fi, err := os.Stat(filepath)
	if err != nil {
		return 0, err
	}
	// get the size
	return fi.Size(), nil
}

func WriteToTempFile(reader io.Reader, tempFilePattern string) (string, error) {
	out, err := os.CreateTemp("", tempFilePattern)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, reader)
	return out.Name(), err
}

func FileExist(fpath string) bool {
	if _, err := os.Stat(fpath); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
