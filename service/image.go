package service

import (
	"errors"

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

func (s imageService) GetInfo(hashid string) (*models.Image, error) {
	i64Arr, err := vars.HashId.DecodeInt64WithError(hashid)
	if err != nil {
		return nil, nil
	}
	var image models.Image
	err = vars.Database.First(&image, i64Arr[0]).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &image, nil
}

func (s imageService) GetCachedInfo(hashid string) (m *models.Image, err error) {
	cacheKey := []byte("imginfo" + hashid)

	if val, err := vars.MemoryCache.Get(cacheKey); err == nil {
		err = sonic.Unmarshal(val, &m)
		if err != nil {
			logrus.Errorln(err)
		}
		return m, err
	}

	content, err, _ := imageInfoSf.Do(hashid, func() (interface{}, error) {
		return s.GetInfo(hashid)
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
	return vars.Database.Delete(&models.Image{}, id).Error
}

func (s imageService) CountHash(hash string) (count int64, err error) {
	err = vars.Database.Model(&models.Image{}).Where("file_hash = ?", hash).Count(&count).Error
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
