package store

import (
	"errors"

	"github.com/RocketChat/filestore-migrator/v2/pkg/models"
)

var (
	// ErrNotFound is returned when a file is not found
	ErrNotFound = errors.New("not found")
)

// Provider describes the basic contract provided to access a static content storage provider.
type Provider interface {
	// StoreType returns the name of the store
	StoreType() string
	// Upload uploads a file from given path to the storage provider
	Upload(objectPath string, filePath string, contentType string) error
	// Download downloads a file from the storage provider and moves it to the temporary file store
	Download(fileCollection string, file models.RocketChatFile) (string, error)
	// SetTempDirectory allows for the setting of the directory that will be used for temporary file store during operations
	SetTempDirectory(subdir string)

	Delete(file models.RocketChatFile, permanentelyDelete bool) error
}
