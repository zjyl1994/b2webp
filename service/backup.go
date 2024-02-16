package service

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/zjyl1994/b2webp/common/utils"
	"github.com/zjyl1994/b2webp/common/vars"
)

func BackupDatabase() error {
	backupFile := filepath.Join(vars.DataDir, fmt.Sprintf("backup.%s.db", time.Now().Format("060102150405")))

	err := vars.Database.Exec("VACUUM INTO ?", backupFile).Error
	if err != nil {
		return err
	}
	defer os.Remove(backupFile)

	contentMD5, err := utils.CalcFileContentMD5(backupFile)
	if err != nil {
		return err
	}

	return S3Service.Put(backupFile, "backup.db", "application/x-sqlite3", contentMD5)
}
