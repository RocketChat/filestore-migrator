package fileStores

import "github.com/RocketChat/MigrateFileStore/models"

type GoogleStorage struct {
}

func (g *GoogleStorage) StoreType() string {
	return "GoogleStorage"
}

func (g *GoogleStorage) Download(fileCollection string, file models.File) (string, error) {

	return "", nil
}

func (g *GoogleStorage) Upload(path string, filePath string, contentType string) error {
	return nil
}
