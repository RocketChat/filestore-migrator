package fileStores

import (
	"github.com/RocketChat/MigrateFileStore/models"
)

type FileStore interface {
	StoreType() string
	Upload(objectPath string, filePath string, contentType string) error
	Download(fileCollection string, file models.File) (string, error)
}
