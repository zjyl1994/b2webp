package server

import (
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/zjyl1994/b2webp/common/models"
	"github.com/zjyl1994/b2webp/common/utils"
	"github.com/zjyl1994/b2webp/common/vars"
	"github.com/zjyl1994/b2webp/service"
	"golang.org/x/sync/singleflight"
)

var s3fetch singleflight.Group

const (
	IMAGE_CACHE_CONTROL = "public, max-age=604800"
)

func UploadImagePage(c *fiber.Ctx) error {
	return c.Render("upload", fiber.Map{
		"motd":          template.HTML(strings.ReplaceAll(vars.Motd, `\n`, "<br>")),
		"total_count":   vars.TotalImageCount,
		"total_size":    vars.TotalImageSize,
		"need_password": len(vars.UploadPassword) > 0,
	})
}

func UploadImageHandler(c *fiber.Ctx) error {
	if vars.UploadPassword != c.FormValue("password") {
		return jsonResult(c, nil, errors.New("上传密码不正确"))
	}

	fh, err := c.FormFile("image")
	if err != nil {
		return jsonResult(c, nil, err)
	}

	contentMD5, err := utils.CalcMultipartFileHeaderContentMD5(fh)
	if err != nil {
		return jsonResult(c, nil, err)
	}
	fileName := utils.Base64ToUrlSafe(contentMD5)
	contentType := fh.Header.Get("Content-Type")
	if !utils.Contains(contentType, []string{"image/jpeg", "image/gif", "image/png", "image/webp"}) {
		return jsonResult(c, nil, errors.New("不支持的文件类型"))
	}

	sameFileCount, err := service.ImageService.CountHash(contentMD5)
	if err != nil {
		return jsonResult(c, nil, err)
	}

	if sameFileCount == 0 {
		cacheFile := service.FileCacheService.GetPath(fileName)

		if err = c.SaveFile(fh, cacheFile); err != nil {
			return jsonResult(c, nil, err)
		}

		if err = service.S3Service.Put(cacheFile, fileName, contentType, contentMD5); err != nil {
			return jsonResult(c, nil, err)
		}
	}

	image, err := service.ImageService.Create(contentType, contentMD5, uint64(fh.Size))
	if err != nil {
		return jsonResult(c, nil, err)
	}

	atomic.AddInt64(&vars.TotalImageCount, 1)
	atomic.AddInt64(&vars.TotalImageSize, fh.Size)

	renderItem, err := image2RenderItem(image)
	if err != nil {
		return jsonResult(c, nil, err)
	}
	return jsonResult(c, renderItem, nil)
}

func DeleteImageHandler(c *fiber.Ctx) error {
	imageId := parseHashId(c.Params("hashid"))
	if imageId == 0 {
		return fiber.ErrNotFound
	}
	image, err := service.ImageService.GetCachedInfo(imageId)
	if err != nil {
		return err
	}
	if image == nil {
		return fiber.ErrNotFound
	}

	delCode := c.Params("delcode")
	if image.DeleteCode != delCode {
		return fiber.ErrForbidden
	}

	err = service.ImageService.Delete(image.ID)
	if err != nil {
		return err
	}

	count, err := service.ImageService.CountHash(image.FileHash)
	if err != nil {
		return err
	}

	if count == 0 {
		fileName := utils.Base64ToUrlSafe(image.FileHash)

		if err = os.Remove(service.FileCacheService.GetPath(fileName)); err != nil {
			return err
		}
		if err = service.S3Service.Delete(fileName); err != nil {
			return err
		}
	}
	atomic.AddInt64(&vars.TotalImageCount, -1)
	atomic.AddInt64(&vars.TotalImageSize, -int64(image.FileSize))
	return fiber.NewError(fiber.StatusOK, "图片已成功删除")
}

func GetImagePage(c *fiber.Ctx) error {
	imageId := parseHashId(c.Params("hashid"))
	if imageId == 0 {
		return fiber.ErrNotFound
	}
	image, err := service.ImageService.GetCachedInfo(imageId)
	if err != nil {
		return err
	}
	if image == nil {
		return fiber.ErrNotFound
	}

	renderItem, err := image2RenderItem(image)
	if err != nil {
		return err
	}
	return c.Render("info", fiber.Map{
		"info": renderItem,
		"img":  template.URL(renderItem.ImageURL),
	})
}

func GetImageHandler(c *fiber.Ctx) error {
	hashId := c.Params("hashid")
	imageId := parseHashId(hashId)
	if imageId == 0 {
		return fiber.ErrNotFound
	}
	image, err := service.ImageService.GetCachedInfo(imageId)
	if err != nil {
		return err
	}
	if image == nil {
		return fiber.ErrNotFound
	}

	fileName := utils.Base64ToUrlSafe(image.FileHash)
	cacheFile := service.FileCacheService.GetPath(fileName)
	if utils.FileExist(cacheFile) {
		return sendImage(c, image, cacheFile)
	}

	_, err, _ = s3fetch.Do(hashId, func() (interface{}, error) {
		return nil, service.S3Service.Get(fileName, cacheFile)
	})
	if err != nil {
		return err
	}

	return sendImage(c, image, cacheFile)
}

func sendImage(c *fiber.Ctx, image *models.Image, filename string) error {
	if utils.Contains(image.ContentType, []string{"image/jpeg", "image/png"}) &&
		c.Accepts("image/webp") == "image/webp" {
		webpFile := utils.ChangeExtname(filename, ".webp")
		if !utils.FileExist(webpFile) {
			if err := service.Convert2Webp(filename, webpFile); err != nil {
				logrus.Errorln(err)
				return sendImageFile(c, filename)
			}
		}
		return sendImageFile(c, webpFile)
	}
	return sendImageFile(c, filename)
}

func sendImageFile(c *fiber.Ctx, filename string) error {
	now := time.Now()
	os.Chtimes(filename, now, now)
	c.Set(fiber.HeaderCacheControl, IMAGE_CACHE_CONTROL)
	return c.SendFile(filename)
}

type imageRenderItem struct {
	ImageURL    string `json:"image_url"`
	InfoPage    string `json:"info_page"`
	DeleteLink  string `json:"delete_link"`
	FileSize    uint64 `json:"file_size"`
	FileHash    string `json:"file_hash"`
	ContentType string `json:"content_type"`
	UploadTime  int64  `json:"upload_time"`
	HashId      string `json:"hash_id"`
}

func image2RenderItem(image *models.Image) (imageRenderItem, error) {
	hashid, err := vars.HashId.EncodeInt64([]int64{int64(image.ID)})
	if err != nil {
		return imageRenderItem{}, err
	}

	baseUrl, err := url.Parse(vars.BaseURL)
	if err != nil {
		return imageRenderItem{}, err
	}
	var extName string
	switch image.ContentType {
	case "image/webp":
		extName = ".webp"
	case "image/gif":
		extName = ".gif"
	case "image/jpeg":
		extName = ".jpg"
	case "image/png":
		extName = ".png"
	}
	baseUrl.Path = fmt.Sprintf("/%s%s", hashid, extName)
	imageUrl := baseUrl.String()

	baseUrl.Path = fmt.Sprintf("/delete/%s/%s", hashid, image.DeleteCode)
	deleteLink := baseUrl.String()

	baseUrl.Path = fmt.Sprintf("/info/%s", hashid)
	infoPage := baseUrl.String()

	return imageRenderItem{
		ImageURL:    imageUrl,
		InfoPage:    infoPage,
		HashId:      hashid,
		DeleteLink:  deleteLink,
		FileSize:    image.FileSize,
		ContentType: image.ContentType,
		UploadTime:  image.UploadTime,
		FileHash:    image.FileHash,
	}, nil
}

func jsonResult(c *fiber.Ctx, data any, err error) error {
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "error": err.Error()})
	} else {
		return c.JSON(fiber.Map{"success": true, "data": data})
	}
}

func parseHashId(s string) uint64 {
	hashid := utils.BareFilename(s)
	i64Arr, err := vars.HashId.DecodeInt64WithError(hashid)
	if err != nil {
		return 0
	}
	return uint64(i64Arr[0])
}
