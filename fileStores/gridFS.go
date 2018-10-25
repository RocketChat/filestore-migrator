package fileStores

import (
	"fmt"
	"io"
	"os"

	"github.com/RocketChat/MigrateFileStore/models"
	mgo "gopkg.in/mgo.v2"
)

type GridFS struct {
	Database string
	Session  *mgo.Session
}

func (g *GridFS) StoreType() string {
	return "GridFS"
}

func (g *GridFS) Download(fileCollection string, file models.File) (string, error) {
	sess := g.Session.Copy()
	defer sess.Close()

	gridFile, err := sess.DB(g.Database).GridFS(fileCollection).Open(file.ID)
	if err != nil {
		if err == mgo.ErrNotFound {
			fmt.Printf("File not found in GridFS: %s", file.Name)
			return "", models.ErrNotFound
		}

		return "", err
	}

	path := "files/" + file.ID

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(f, gridFile)
	if err != nil {
		return "", err
	}

	err = f.Close()
	if err != nil {
		return "", err
	}

	err = gridFile.Close()
	if err != nil {
		return "", err
	}

	return path, err
}

func (g *GridFS) Upload(path string, filePath string, contentType string) error {
	return nil
}
