package filestores

import (
	"github.com/RocketChat/MigrateFileStore/models"
)

// FileStore is the file store interface
type FileStore interface {
	StoreType() string
	Upload(objectPath string, filePath string, contentType string) error
	Download(fileCollection string, file models.File) (string, error)
	SetTempDirectory(subdir string)
}
