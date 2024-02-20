package main

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/coocood/freecache"
	"github.com/inhies/go-bytesize"
	_ "github.com/joho/godotenv/autoload"
	gorm_logrus "github.com/onrik/gorm-logrus"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/speps/go-hashids"
	"github.com/zjyl1994/b2webp/common/models"
	"github.com/zjyl1994/b2webp/common/utils"
	"github.com/zjyl1994/b2webp/common/vars"
	"github.com/zjyl1994/b2webp/server"
	"github.com/zjyl1994/b2webp/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	err := errMain()
	if err != nil {
		logrus.Fatalln(err.Error())
	}
}

func errMain() (err error) {
	vars.BootTime = time.Now()

	if !utils.CmdExist("cwebp") {
		return errors.New("missing cwebp command")
	}

	vars.DebugMode, err = strconv.ParseBool(vars.Getenv("B2WEBP_DEBUG"))
	if vars.DebugMode {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debugln("B2WEBP will run in debug mode.")
	}

	vars.ListenAddress = vars.Getenv("B2WEBP_LISTEN")
	logrus.Debugln("Listen address", vars.ListenAddress)

	vars.DataDir = vars.Getenv("B2WEBP_DATA_PATH")
	logrus.Debugln("Data path", vars.DataDir)

	vars.BaseURL = utils.Coalesce(vars.Getenv("B2WEBP_BASE_URL"), "http://"+vars.ListenAddress+"/")
	if err != nil {
		return err
	}
	logrus.Debugln("Base url", vars.BaseURL)

	vars.CacheDir = filepath.Join(vars.DataDir, "cache")
	err = os.MkdirAll(vars.CacheDir, 0755)
	if err != nil {
		return err
	}
	logrus.Debugln("Cache path", vars.CacheDir)

	vars.CdnAssetsPrefix = vars.Getenv("B2WEBP_CDN_ASSETS_PREFIX")
	vars.Motd = vars.Getenv("B2WEBP_MOTD")

	if maxCacheSize, err := bytesize.Parse(strings.ToUpper(vars.Getenv("B2WEBP_S3_MAX_CACHE_SIZE"))); err != nil {
		return err
	} else {
		vars.MaxCacheSize = int64(maxCacheSize)
		logrus.Debugln("Max disk cache size", maxCacheSize.String())
	}

	if memoryCacheSize, err := bytesize.Parse(strings.ToUpper(vars.Getenv("B2WEBP_MEMORY_CACHE_SIZE"))); err != nil {
		return err
	} else {
		vars.MemoryCache = freecache.NewCache(int(memoryCacheSize))
		logrus.Debugln("Memory cache size", memoryCacheSize.String())
	}

	hashidSetting := hashids.HashIDData{Alphabet: hashids.DefaultAlphabet, Salt: vars.Getenv("B2WEBP_HASHID_SALT"), MinLength: 6}
	logrus.Debugln("HashID setting", hashidSetting)
	vars.HashId, err = hashids.NewWithData(&hashidSetting)
	if err != nil {
		return err
	}

	if val := vars.Getenv("B2WEBP_UPLOAD_PASSWORD"); len(val) > 0 {
		vars.UploadPassword = val
		logrus.Infoln("Upload password enabled with environment variables 'B2WEBP_UPLOAD_PASSWORD'.")
	}

	vars.CronInstance = cron.New()
	vars.CronInstance.AddFunc("@daily", func() {
		if err := service.FileCacheService.Clean(); err != nil {
			logrus.Errorln(err)
		}
		logrus.Infoln("disk cache cleaned.")
		if err := service.BackupDatabase(); err != nil {
			logrus.Errorln(err)
		}
		logrus.Infoln("database backup uploaded.")
	})
	vars.CronInstance.Start()

	vars.S3Setting = vars.S3SettingS{
		Region:       vars.Getenv("B2WEBP_S3_REGION"),
		Endpoint:     vars.Getenv("B2WEBP_S3_ENDPOINT"),
		Bucket:       vars.Getenv("B2WEBP_S3_BUCKET"),
		AccessId:     vars.Getenv("B2WEBP_S3_ACCESS_ID"),
		AccessKey:    vars.Getenv("B2WEBP_S3_ACCESS_KEY"),
		ObjectPrefix: vars.Getenv("B2WEBP_S3_OBJECT_PREFIX"),
	}

	databasePath := filepath.Join(vars.DataDir, "b2webp.db")
	logrus.Debugln("Database path", databasePath)
	vars.Database, err = gorm.Open(sqlite.Open(databasePath), &gorm.Config{
		Logger: gorm_logrus.New(),
	})
	if err != nil {
		return err
	}

	err = vars.Database.AutoMigrate(&models.Image{})
	if err != nil {
		return err
	}

	logrus.Infoln("Starting B2WEBP service...")
	return server.Run(vars.ListenAddress)
}
