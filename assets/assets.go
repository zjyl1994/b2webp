package assets

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/zjyl1994/b2webp/common/vars"
)

//go:embed template static
var assetsFS embed.FS

//go:embed favicon.ico
var Favicon []byte

func GetFS(dir string) http.FileSystem {
	subfs, err := fs.Sub(assetsFS, dir)
	if err != nil {
		panic(err)
	}
	diskPath := filepath.Join(vars.Getenv("B2WEBP_ASSETS_PATH"), dir)
	if stat, err := os.Stat(diskPath); err == nil && stat.IsDir() {
		subfs = os.DirFS(diskPath)
	}
	return http.FS(subfs)
}
