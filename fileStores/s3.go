package fileStores

import (
	"fmt"
	"io"
	"os"

	"github.com/RocketChat/MigrateFileStore/models"
	minio "github.com/minio/minio-go"
)

type S3 struct {
	Endpoint         string
	Bucket           string
	AccessID         string
	AccessKey        string
	Region           string
	UseSSL           bool
	TempFileLocation string
}

func (s *S3) StoreType() string {
	return "AmazonS3"
}

func (s *S3) Download(fileCollection string, file models.File) (string, error) {
	minioClient, err := minio.NewWithRegion(s.Endpoint, s.AccessID, s.AccessKey, s.UseSSL, s.Region)
	if err != nil {
		return "", err
	}

	filePath := s.TempFileLocation + "/" + file.ID

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

	return filePath, nil
}

func (s *S3) Upload(objectPath string, filePath string, contentType string) error {
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
