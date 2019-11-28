package fileStores

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/RocketChat/MigrateFileStore/models"
	mgo "gopkg.in/mgo.v2"
)

type GridFS struct {
	Database         string
	Session          *mgo.Session
	TempFileLocation string
}

func (g *GridFS) StoreType() string {
	return "GridFS"
}

func (g *GridFS) SetTempDirectory(dir string) {
	g.TempFileLocation = dir
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

	defer gridFile.Close()

	filePath := g.TempFileLocation + "/" + file.ID

	log.Println(filePath)

	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}

	defer f.Close()

	if _, err = io.Copy(f, gridFile); err != nil {
		return "", err
	}

	return filePath, err
}

func (g *GridFS) Upload(path string, filePath string, contentType string) error {
	return nil
}
