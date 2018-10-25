package fileStores

import (
	"fmt"

	"github.com/RocketChat/MigrateFileStore/models"
	minio "github.com/minio/minio-go"
)

type S3 struct {
	Endpoint  string
	Bucket    string
	AccessId  string
	AccessKey string
	Region    string
	UseSSL    bool
}

func (s *S3) StoreType() string {
	return "AmazonS3"
}

func (s *S3) Download(fileCollection string, file models.File) (string, error) {

	return "", nil
}

func (s *S3) Upload(path string, filePath string, contentType string) error {
	minioClient, err := minio.NewWithRegion(s.Endpoint, s.AccessId, s.AccessKey, s.UseSSL, s.Region)
	if err != nil {
		return err
	}

	_, err = minioClient.FPutObject(s.Bucket, path, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
