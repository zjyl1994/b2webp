package vars

const (
	APP_NAME = "B2WEBP"
)

type S3SettingS struct {
	Region       string
	Endpoint     string
	Bucket       string
	AccessId     string
	AccessKey    string
	ObjectPrefix string
}

var defaultEnvVar = map[string]string{
	"B2WEBP_SITE_NAME":         APP_NAME,
	"B2WEBP_DEBUG":             "false",
	"B2WEBP_LISTEN":            "127.0.0.1:9000",
	"B2WEBP_DATA_PATH":         "./data",
	"B2WEBP_HASHID_SALT":       APP_NAME,
	"B2WEBP_S3_MAX_CACHE_SIZE": "500MB",
	"B2WEBP_MEMORY_CACHE_SIZE": "50MB",
	"B2WEBP_ASSETS_PATH":       "assets",
	"B2WEBP_GUEST_UPLOAD":      "true",
}
