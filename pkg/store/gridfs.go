package store

import (
	"errors"
	"os"

	"github.com/RocketChat/filestore-migrator/v2/pkg/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GridFSProvider provides methods to use GridFS as a storage provider.
type GridFSProvider struct {
	Database         string
	Session          mongo.Session
	TempFileLocation string

	Buckets map[string]*gridfs.Bucket
}

// StoreType returns the name of the store
func (g *GridFSProvider) StoreType() string {
	return "GridFS"
}

func (g *GridFSProvider) addBucket(bucketName string) (*gridfs.Bucket, error) {
	if bucket, err := gridfs.NewBucket(g.Session.Client().Database(g.Database), options.GridFSBucket().SetName(bucketName)); err != nil {
		return nil, err
	} else {
		g.Buckets[bucketName] = bucket
	}

	return g.Buckets[bucketName], nil
}

// SetTempDirectory allows for the setting of the directory that will be used for temporary file store during operations
func (g *GridFSProvider) SetTempDirectory(dir string) {
	g.TempFileLocation = dir
}

// Download downloads a file from the storage provider and moves it to the temporary file store
func (g *GridFSProvider) Download(fileCollection string, file models.RocketChatFile) (string, error) {

	var (
		bucket *gridfs.Bucket
		ok     bool
		err    error
	)
	if bucket, ok = g.Buckets[fileCollection]; !ok {
		bucket, err = g.addBucket(fileCollection)
		if err != nil {
			return "", err
		}
	}

	filePath := g.TempFileLocation + "/" + file.ID

	if _, err = os.Stat(filePath); os.IsNotExist(err) {

		f, err := os.Create(filePath)
		if err != nil {
			return "", err
		}

		if _, err = bucket.DownloadToStream(file.ID, f); err != nil {
			return "", err
		}

		f.Close()
	}

	return filePath, nil
}

// Upload uploads a file from given path to the storage provider (not implemented)
func (g *GridFSProvider) Upload(path string, filePath string, contentType string) error {
	return errors.New("unimplemented")
}

func (s *GridFSProvider) Delete(file models.RocketChatFile, permanentelyDelete bool) error {
	return errors.New("unimplemented")
}
