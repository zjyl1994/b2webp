package models

type Image struct {
	ID         uint64 `gorm:"primaryKey"`
	UploadTime int64  `gorm:"autoCreateTime"`

	ContentType string
	FileHash    string
	FileSize    uint64

	DeleteCode string
}
