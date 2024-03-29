package vars

import (
	"os"
	"sync/atomic"
	"time"

	"github.com/coocood/freecache"
	"github.com/robfig/cron/v3"
	"github.com/speps/go-hashids"
	"gorm.io/gorm"
)

var (
	ListenAddress string
	DebugMode     bool
	DataDir       string
	BaseURL       string
	BootTime      time.Time
	CacheDir      string
	MaxCacheSize  int64

	S3Setting S3SettingS

	Database     *gorm.DB
	HashId       *hashids.HashID
	MemoryCache  *freecache.Cache
	CronInstance *cron.Cron

	UploadPassword    string
	CdnAssetsPrefix   string
	Motd              string
	ServeByteCounter  atomic.Uint64
	ServeClickCounter atomic.Uint64
)

func Getenv(name string) string {
	val := os.Getenv(name)
	if len(val) > 0 {
		return val
	}
	if val, ok := defaultEnvVar[name]; ok {
		return val
	}
	return ""
}
