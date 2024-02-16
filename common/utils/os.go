package utils

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CopyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	fin, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fin.Close()
	fout, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fout.Close()
	_, err = io.Copy(fout, fin)
	return err
}

func FileExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return false
}

type FileInfo struct {
	Path  string
	Size  int64
	Mtime int64
}

func ScanFileInDir(dirPath string) ([]FileInfo, error) {
	items, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	list := make([]FileInfo, 0, len(items))
	for _, item := range items {
		if item.IsDir() {
			continue
		}
		info, err := item.Info()
		if err != nil {
			return nil, err
		}
		list = append(list, FileInfo{
			Path:  filepath.Join(dirPath, item.Name()),
			Size:  info.Size(),
			Mtime: info.ModTime().Unix(),
		})
	}
	return list, nil
}

func BareFilename(filename string) string {
	return strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
}

func ChangeExtname(filename, extname string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename)) + extname
}

func CmdExist(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
