package fileStores

import (
	"io"
	"os"

	"github.com/RocketChat/MigrateFileStore/models"
	mgo "gopkg.in/mgo.v2"
)

// GridFS is the GridFS file store
type GridFS struct {
	Database         string
	Session          *mgo.Session
	TempFileLocation string
}

// StoreType returns the name of the store
func (g *GridFS) StoreType() string {
	return "GridFS"
}

// SetTempDirectory allows for the setting of the directory that will be used for temporary file store during operations
func (g *GridFS) SetTempDirectory(dir string) {
	g.TempFileLocation = dir
}

// Download will download the file to temp file store
func (g *GridFS) Download(fileCollection string, file models.File) (string, error) {
	sess := g.Session.Copy()
	defer sess.Close()

	gridFile, err := sess.DB(g.Database).GridFS(fileCollection).Open(file.ID)
	if err != nil {
		if err == mgo.ErrNotFound {
			return "", models.ErrNotFound
		}

		return "", err
	}

	defer gridFile.Close()

	filePath := g.TempFileLocation + "/" + file.ID

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

// Upload will upload the file from given file path - not implemented
func (g *GridFS) Upload(path string, filePath string, contentType string) error {
	return nil
}
