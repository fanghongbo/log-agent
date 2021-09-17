package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
)

func GetStringMd5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func GetFileMd5(filePath string) (string, error) {
	var (
		file    *os.File
		hashVal hash.Hash
		err     error
	)

	if file, err = os.Open(filePath); err != nil {
		return "", err
	}

	hashVal = md5.New()
	if _, err = io.Copy(hashVal, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hashVal.Sum([]byte(""))), nil
}
