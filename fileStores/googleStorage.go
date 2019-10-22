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

type GoogleStorage struct {
	JSONKey          string
	Bucket           string
	TempFileLocation string
}

func (g *GoogleStorage) StoreType() string {
	return "GoogleCloudStorage"
}

func (g *GoogleStorage) Download(fileCollection string, file models.File) (string, error) {

	ctx := context.Background()

	cfg, err := google.JWTConfigFromJSON([]byte(g.JSONKey), "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", err
	}

	c := cfg.Client(ctx)

	service, err := storage.New(c)

	filePath := g.TempFileLocation + "/" + file.ID

	log.Println("Downloading", file.GoogleStorage.Path)

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

	log.Println("Downloaded", file.GoogleStorage.Path)

	return filePath, nil
}

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

	obj, err := insertCall.Do()
	if err != nil {
		log.Println(err)
		return errors.New("problem uploading file to bucket")
	}

	log.Println(obj)

	return nil
}
