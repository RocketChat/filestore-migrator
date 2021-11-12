package store

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/RocketChat/filestore-migrator/rocketchat"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/tags"
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
	minioClient, err := minio.New(s.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.AccessID, s.AccessKey, ""),
		Secure: s.UseSSL,
		Region: s.Region,
	})
	if err != nil {
		return "", err
	}

	filePath := s.TempFileLocation + "/" + file.ID

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		object, err := minioClient.GetObject(
			context.Background(),
			s.Bucket,
			file.AmazonS3.Path,
			minio.GetObjectOptions{},
		)
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
	minioClient, err := minio.New(s.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.AccessID, s.AccessKey, ""),
		Secure: s.UseSSL,
		Region: s.Region,
	})
	if err != nil {
		return err
	}

	_, err = minioClient.FPutObject(
		context.Background(),
		s.Bucket,
		objectPath,
		filePath,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// Delete permanentely permanentely destroys an object specified by the
// rocketFile.Amazons3.filepath
func (s *S3Provider) Delete(file rocketchat.File, permanentelyDelete bool) error {
	minioClient, err := minio.New(s.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.AccessID, s.AccessKey, ""),
		Secure: s.UseSSL,
		Region: s.Region,
	})
	if err != nil {
		return err
	}

	// removes the bucket name from the Path if it exists
	objectPrefix := strings.TrimPrefix(file.AmazonS3.Path, s.Bucket)

	// chan of objects withing the deployment object
	objectsCh := make(chan string)

	//send object names thata are to be removed to objectsCh
	go func() {
		defer close(objectsCh)
		recursive := true

		for object := range minioClient.ListObjects(
			context.Background(),
			s.Bucket,
			minio.ListObjectsOptions{
				Prefix:    objectPrefix,
				Recursive: recursive,
			},
		) {
			if object.Err != nil {
				log.Println(object.Err)
			}

			objectsCh <- object.Key
		}
	}()

	// tags that will mark objects for deletion
	objTags, err := tags.NewTags(map[string]string{"delete": "true"}, true)
	if err != nil {
		return err
	}

	if permanentelyDelete {
		log.Println("permanentely deleting all the objects of the deployment")
		for objName := range objectsCh {
			err := minioClient.RemoveObject(
				context.Background(),
				s.Bucket,
				objName,
				minio.RemoveObjectOptions{})
			if err != nil {
				return err
			}

		}
		log.Println("permanentely deleting the deployment object itself")
		return minioClient.PutObjectTagging(
			context.Background(),
			s.Bucket,
			file.AmazonS3.Path,
			objTags,
			// specifies version of the object
			minio.PutObjectTaggingOptions{VersionID: ""},
		)

	}

	// execute only if !permanentelyDelete
	log.Println("marking all the objects of the deployment for deletion")
	for objName := range objectsCh {
		err := minioClient.PutObjectTagging(
			context.Background(),
			s.Bucket,
			objName,
			objTags,
			// specifies version of the object, not used yet
			minio.PutObjectTaggingOptions{VersionID: ""},
		)
		if err != nil {
			return err
		}
	}

	log.Println("Deleting the deployment object itself")
	return minioClient.PutObjectTagging(
		context.Background(),
		s.Bucket,
		file.AmazonS3.Path,
		objTags,
		// specifies version of the object
		minio.PutObjectTaggingOptions{VersionID: ""},
	)
}
