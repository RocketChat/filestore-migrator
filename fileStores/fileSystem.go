package fileStores

import (
	"io"
	"os"

	"github.com/RocketChat/MigrateFileStore/models"
)

type FileSystem struct {
	Location string
}

func (f *FileSystem) StoreType() string {
	return "FileSystem"
}

func (f *FileSystem) Download(fileCollection string, file models.File) (string, error) {

	sourcePath := f.Location + "/" + file.ID
	destinationPath := "files/" + file.ID

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
