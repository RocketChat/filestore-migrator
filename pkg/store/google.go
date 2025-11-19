package store

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"strings"

	"github.com/RocketChat/filestore-migrator/v2/pkg/models"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/storage/v1"
)

// GoogleStorageProvider provides methods to use the Google Cloud Storage offering as a storage provider.
type GoogleStorageProvider struct {
	JSONKey          string
	Bucket           string
	TempFileLocation string
}

// StoreType returns the name of the store
func (g *GoogleStorageProvider) StoreType() string {
	return "GoogleCloudStorage"
}

// SetTempDirectory allows for the setting of the directory that will be used for temporary file store during operations
func (g *GoogleStorageProvider) SetTempDirectory(dir string) {
	g.TempFileLocation = dir
}

// Download downloads a file from the storage provider and moves it to the temporary file store
func (g *GoogleStorageProvider) Download(fileCollection string, file models.RocketChatFile) (string, error) {
	ctx := context.Background()

	cfg, err := google.JWTConfigFromJSON([]byte(g.JSONKey), "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", err
	}

	c := cfg.Client(ctx)

	service, err := storage.New(c)

	filePath := g.TempFileLocation + "/" + file.ID

	if _, err := os.Stat(filePath); os.IsNotExist(err) {

		getCall := service.Objects.Get(g.Bucket, file.GoogleStorage.Path)
		resp, err := getCall.Download()
		if err != nil {
			if strings.Contains(err.Error(), "No such object:") {
				return "", ErrNotFound
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
	}

	return filePath, nil
}

// Upload uploads a file from given path to the storage provider
func (g *GoogleStorageProvider) Upload(path string, filePath string, contentType string) error {
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

func (s *GoogleStorageProvider) Delete(file models.RocketChatFile, permanentelyDelete bool) error {
	return errors.New("delete object method not implemented")
}
