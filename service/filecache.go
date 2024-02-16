package service

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/zjyl1994/b2webp/common/utils"
	"github.com/zjyl1994/b2webp/common/vars"
)

var FileCacheService fileCacheService

type fileCacheService struct{}

func (s fileCacheService) GetPath(fileName string) string {
	return filepath.Join(vars.CacheDir, fileName)
}

func (s fileCacheService) Clean() error {
	list, err := utils.ScanFileInDir(vars.CacheDir)
	if err != nil {
		return err
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Mtime < list[j].Mtime })
	var totalCount int64
	for _, v := range list {
		totalCount += v.Size
	}

	if vars.MaxCacheSize > totalCount {
		return nil
	}

	for _, v := range list {
		totalCount -= v.Size
		err = os.Remove(v.Path)
		if err != nil {
			return err
		}
		if vars.MaxCacheSize > totalCount {
			break
		}
	}
	return nil
}
