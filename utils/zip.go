package utils

import (
	"github.com/alexmullins/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Zip(srcFile string, destZip string, password string) error {
	var (
		zipFile *os.File
		file    *os.File
		archive *zip.Writer
		err     error
	)

	if zipFile, err = os.Create(destZip); err != nil {
		return err
	}

	defer func() {
		_ = zipFile.Close()
	}()

	archive = zip.NewWriter(zipFile)

	defer func() {
		_ = archive.Close()
	}()

	if err = filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {
		var (
			header *zip.FileHeader
			w      io.Writer
		)

		if err != nil {
			return err
		}

		if header, err = zip.FileInfoHeader(info); err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(path, filepath.Dir(srcFile)+"/")
		// header.Name = path

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		if password != "" {
			header.SetPassword(password)
		}

		if w, err = archive.CreateHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			if file, err = os.Open(path); err != nil {
				return err
			}

			defer func() {
				_ = file.Close()
			}()

			_, err = io.Copy(w, file)
		}
		return err
	}); err != nil {
		return err
	}

	return nil
}

func Unzip(zipFile string, destDir string, password string) error {
	var (
		zipReader *zip.ReadCloser
		err       error
	)

	if zipReader, err = zip.OpenReader(zipFile); err != nil {
		return err
	}

	defer func() {
		_ = zipReader.Close()
	}()

	for _, item := range zipReader.File {
		var (
			filePath string
			err      error
		)

		if password != "" {
			item.SetPassword(password)
		}

		filePath = filepath.Join(destDir, item.Name)

		if item.FileInfo().IsDir() {
			if err = os.MkdirAll(filePath, os.ModePerm); err != nil {
				return err
			}
		} else {
			if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
				return err
			}

			if err = saveZipFile(item, filePath); err != nil {
				return err
			}
		}
	}
	return nil
}

func saveZipFile(file *zip.File, filePath string) error {
	var (
		src  io.ReadCloser
		dest *os.File
		err  error
	)

	if src, err = file.Open(); err != nil {
		return err
	}

	defer func() {
		_ = src.Close()
	}()

	if dest, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode()); err != nil {
		return err
	}

	defer func() {
		_ = dest.Close()
	}()

	if _, err = io.Copy(dest, src); err != nil {
		return err
	}

	return nil
}
