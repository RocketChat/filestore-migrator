package store

import (
	"fmt"
	"io"
	"os"

	"github.com/RocketChat/filestore-migrator/rocketchat"
	minio "github.com/minio/minio-go"
)

// S3Provider provides methods to use any S3 complaint provider as a storage provider.
type S3Provider struct {
	Endpoint         string
	Bucket           string
	AccessID         string
	AccessKey        string
	Region           string
	UseSSL           bool
	TempFileLocation string
}

// StoreType returns the name of the store
func (s *S3Provider) StoreType() string {
	return "AmazonS3"
}

// SetTempDirectory allows for the setting of the directory that will be used for temporary file store during operations
func (s *S3Provider) SetTempDirectory(dir string) {
	s.TempFileLocation = dir
}

// Download will download the file to temp file store
func (s *S3Provider) Download(fileCollection string, file rocketchat.File) (string, error) {
	minioClient, err := minio.NewWithRegion(s.Endpoint, s.AccessID, s.AccessKey, s.UseSSL, s.Region)
	if err != nil {
		return "", err
	}

	filePath := s.TempFileLocation + "/" + file.ID

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		object, err := minioClient.GetObject(s.Bucket, file.AmazonS3.Path, minio.GetObjectOptions{})
		if err != nil {
			return "", err
		}

		f, err := os.Create(filePath)
		if err != nil {
			return "", err
		}

		defer f.Close()

		if _, err = io.Copy(f, object); err != nil {
			return "", err
		}

	}

	return filePath, nil
}

// Upload will upload the file from given file path
func (s *S3Provider) Upload(objectPath string, filePath string, contentType string) error {
	minioClient, err := minio.NewWithRegion(s.Endpoint, s.AccessID, s.AccessKey, s.UseSSL, s.Region)
	if err != nil {
		return err
	}

	_, err = minioClient.FPutObject(s.Bucket, objectPath, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
