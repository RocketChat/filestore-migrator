package fileStores

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"strings"

	"github.com/RocketChat/MigrateFileStore/models"
	"golang.org/x/oauth2/google"
	storage "google.golang.org/api/storage/v1"
)

// GoogleStorage is the google storage file store
type GoogleStorage struct {
	JSONKey          string
	Bucket           string
	TempFileLocation string
}

// StoreType returns the name of the store
func (g *GoogleStorage) StoreType() string {
	return "GoogleCloudStorage"
}

// SetTempDirectory allows for the setting of the directory that will be used for temporary file store during operations
func (g *GoogleStorage) SetTempDirectory(dir string) {
	g.TempFileLocation = dir
}

// Download will download the file to temp file store
func (g *GoogleStorage) Download(fileCollection string, file models.File) (string, error) {

	ctx := context.Background()

	cfg, err := google.JWTConfigFromJSON([]byte(g.JSONKey), "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", err
	}

	c := cfg.Client(ctx)

	service, err := storage.New(c)

	filePath := g.TempFileLocation + "/" + file.ID

	getCall := service.Objects.Get(g.Bucket, file.GoogleStorage.Path)
	resp, err := getCall.Download()
	if err != nil {
		if strings.Contains(err.Error(), "No such object:") {
			return "", models.ErrNotFound
		}

		return "", err
	}

	defer resp.Body.Close()

	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}

	defer f.Close()

	if _, err = io.Copy(f, resp.Body); err != nil {
		return "", err
	}

	return filePath, nil
}

// Upload will upload the file from given file path
func (g *GoogleStorage) Upload(path string, filePath string, contentType string) error {
	ctx := context.Background()

	cfg, err := google.JWTConfigFromJSON([]byte(g.JSONKey), "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return err
	}

	c := cfg.Client(ctx)

	service, err := storage.New(c)

	file, err := os.Open(filePath)
	if err != nil {
		log.Println(err)
		return errors.New("problem opening file to upload")
	}

	defer file.Close()

	object := &storage.Object{
		Name: path,
	}

	insertCall := service.Objects.Insert(g.Bucket, object).Media(file)

	_, err = insertCall.Do()
	if err != nil {
		log.Println(err)
		return errors.New("problem uploading file to bucket")
	}

	return nil
}
