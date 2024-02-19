package service

import (
	"errors"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/sirupsen/logrus"
	"github.com/zjyl1994/b2webp/common/models"
	"github.com/zjyl1994/b2webp/common/utils"
	"github.com/zjyl1994/b2webp/common/vars"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var imageInfoSf singleflight.Group

const IMAGE_INFO_CACHE_TTL = 3

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
	cacheKey := []byte("imginfo" + strconv.FormatUint(id, 10))

	if val, err := vars.MemoryCache.Get(cacheKey); err == nil {
		err = sonic.Unmarshal(val, &m)
		if err != nil {
			logrus.Errorln(err)
		}
		return m, err
	}

	content, err, _ := imageInfoSf.Do(string(cacheKey), func() (interface{}, error) {
		return s.GetInfo(id)
	})
	if err != nil {
		return nil, err
	}
	m = content.(*models.Image)

	cacheData, err := sonic.Marshal(m)
	if err == nil {
		if err = vars.MemoryCache.Set(cacheKey, cacheData, IMAGE_INFO_CACHE_TTL); err != nil {
			logrus.Errorln(err)
		}
	} else {
		logrus.Errorln(err)
	}
	return m, nil
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
