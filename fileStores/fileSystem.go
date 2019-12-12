package fileStores

import (
	"io"
	"os"

	"github.com/RocketChat/MigrateFileStore/models"
)

// FileSystem is the Filesystem file Store
type FileSystem struct {
	Location         string
	TempFileLocation string
}

// StoreType returns the name of the store
func (f *FileSystem) StoreType() string {
	return "FileSystem"
}

// SetTempDirectory allows for the setting of the directory that will be used for temporary file store during operations
func (f *FileSystem) SetTempDirectory(dir string) {
	f.TempFileLocation = dir
}

// Download will download the file to temp file store
func (f *FileSystem) Download(fileCollection string, file models.File) (string, error) {

	sourcePath := f.Location + "/" + file.ID
	destinationPath := f.TempFileLocation + "/" + file.ID

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return "", models.ErrNotFound
	}

	sF, err := os.Open(sourcePath)
	if err != nil {
		return "", err
	}

	defer sF.Close()

	dF, err := os.Create(destinationPath)
	if err != nil {
		return "", err
	}

	defer dF.Close()

	if _, err = io.Copy(dF, sF); err != nil {
		return "", err
	}

	return destinationPath, nil
}

// Upload will upload the file from given file path
func (f *FileSystem) Upload(path string, filePath string, contentType string) error {
	destinationPath := f.Location + "/" + path

	sF, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer sF.Close()

	dF, err := os.Create(destinationPath)
	if err != nil {
		return err
	}

	defer dF.Close()

	if _, err = io.Copy(dF, sF); err != nil {
		return err
	}

	return nil
}
