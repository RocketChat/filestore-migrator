package store

import (
	"io"
	"os"

	"github.com/RocketChat/filestore-migrator/rocketchat"
)

// FileSystemStorageProvider provides methods to use the local file system as a storage provider.
type FileSystemStorageProvider struct {
	Location         string
	TempFileLocation string
}

// StoreType returns the name of the store
func (f *FileSystemStorageProvider) StoreType() string {
	return "FileSystem"
}

// SetTempDirectory allows for the setting of the directory that will be used for temporary file store during operations
func (f *FileSystemStorageProvider) SetTempDirectory(dir string) {
	f.TempFileLocation = dir
}

// Download downloads a file from the storage provider and moves it to the temporary file store
func (f *FileSystemStorageProvider) Download(fileCollection string, file rocketchat.File) (string, error) {
	sourcePath := f.Location + "/" + file.ID
	destinationPath := f.TempFileLocation + "/" + file.ID

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return "", ErrNotFound
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

// Upload uploads a file from given path to the storage provider
func (f *FileSystemStorageProvider) Upload(path string, filePath string, contentType string) error {
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
