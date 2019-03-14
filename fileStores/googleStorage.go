package fileStores

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/RocketChat/MigrateFileStore/models"
	"golang.org/x/oauth2/google"
	storage "google.golang.org/api/storage/v1"
)

type GoogleStorage struct {
	JSONKey string
	Bucket  string
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

	path := "files/" + file.ID

	log.Println("Downloading", file.GoogleStorage.Path)

	getCall := service.Objects.Get(g.Bucket, file.GoogleStorage.Path)
	resp, err := getCall.Download()
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		panic(err)
	}

	log.Println("Downloaded", file.GoogleStorage.Path)

	return path, nil
}

func (g *GoogleStorage) Upload(path string, filePath string, contentType string) error {
	return nil
}
