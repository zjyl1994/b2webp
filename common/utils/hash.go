package utils

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"mime/multipart"
	"os"
	"strings"
)

func CalcContentMD5(reader io.Reader) (string, error) {
	h := md5.New()
	_, err := io.Copy(h, reader)
	if err != nil {
		return "", err
	}
	result := h.Sum(nil)
	return base64.RawStdEncoding.EncodeToString(result), nil
}

func CalcFileContentMD5(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return CalcContentMD5(f)
}

func CalcMultipartFileHeaderContentMD5(fh *multipart.FileHeader) (string, error) {
	f, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	return CalcContentMD5(f)
}

func Base64ToUrlSafe(b64 string) string {
	b64 = strings.ReplaceAll(b64, "+", "-")
	b64 = strings.ReplaceAll(b64, "/", "_")
	return b64
}
