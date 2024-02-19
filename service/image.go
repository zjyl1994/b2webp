package service

import (
	"errors"
	"strconv"

	"github.com/zjyl1994/b2webp/common/models"
	"github.com/zjyl1994/b2webp/common/utils"
	"github.com/zjyl1994/b2webp/common/vars"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var imageInfoSf singleflight.Group

const IMAGE_INFO_CACHE_TTL = 5

var ImageService imageService

type imageService struct{}

func (s imageService) GetInfo(id uint64) (*models.Image, error) {
	var image models.Image
	err := vars.Database.First(&image, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &image, nil
}

func (s imageService) GetCachedInfo(id uint64) (m *models.Image, err error) {
	return utils.CacheGet(vars.MemoryCache, &imageInfoSf, "imginfo"+strconv.FormatUint(id, 10), func() (*models.Image, error) {
		return s.GetInfo(id)
	}, IMAGE_INFO_CACHE_TTL)
}

func (s imageService) Delete(id uint64) error {
	err := vars.Database.Delete(&models.Image{}, id).Error
	if err != nil {
		return err
	}
	cacheKey := []byte("imginfo" + strconv.FormatUint(id, 10))
	vars.MemoryCache.Del(cacheKey)
	return nil
}

func (s imageService) CountHash(fileHash string) (count int64, err error) {
	err = vars.Database.Model(&models.Image{}).Where("file_hash = ?", fileHash).Count(&count).Error
	return count, err
}

func (s imageService) Create(contentType, fileHash string, fileSize uint64) (*models.Image, error) {
	m := models.Image{
		ContentType: contentType,
		FileHash:    fileHash,
		FileSize:    fileSize,
		DeleteCode:  utils.RandString(8),
	}
	err := vars.Database.Create(&m).Error
	return &m, err
}

func (s imageService) TotalCount() (count int64, err error) {
	err = vars.Database.Model(&models.Image{}).Count(&count).Error
	return count, err
}

func (s imageService) RealTotalCount() (count int64, err error) {
	err = vars.Database.Model(&models.Image{}).Group("file_hash").Count(&count).Error
	return count, err
}

func (s imageService) TotalSize() (sum int64, err error) {
	var result struct {
		FileSize int64
	}
	err = vars.Database.Model(&models.Image{}).Select("SUM(file_size) as file_size").First(&result).Error
	return result.FileSize, err
}

func (s imageService) RealTotalSize() (count int64, err error) {
	var result struct {
		FileSize int64
	}
	err = vars.Database.Table("(?) as u", vars.Database.Model(&models.Image{}).Select("file_size").
		Group("file_hash")).Select("SUM(file_size) as file_size").First(&result).Error
	return result.FileSize, err
}

type StatInfo struct {
	TotalCount        int64
	RealTotalCount    int64
	TotalFileSize     int64
	RealTotalFileSize int64
}

func (s imageService) GetCachedStat() (stat *StatInfo, err error) {
	return utils.CacheGet(vars.MemoryCache, &imageInfoSf, "statinfo", func() (result *StatInfo, err error) {
		result = new(StatInfo)
		result.TotalCount, err = s.TotalCount()
		if err != nil {
			return nil, err
		}
		result.TotalFileSize, err = s.TotalSize()
		if err != nil {
			return nil, err
		}
		result.RealTotalCount, err = s.RealTotalCount()
		if err != nil {
			return nil, err
		}
		result.RealTotalFileSize, err = s.RealTotalSize()
		if err != nil {
			return nil, err
		}
		return result, nil
	}, IMAGE_INFO_CACHE_TTL)
}
