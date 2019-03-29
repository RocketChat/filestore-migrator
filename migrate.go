package MigrateFileStore

import (
	"errors"
	"log"

	"github.com/RocketChat/MigrateFileStore/config"
	"github.com/RocketChat/MigrateFileStore/fileStores"
	mgo "gopkg.in/mgo.v2"
)

var sourceStore fileStores.FileStore
var destinationStore fileStores.FileStore

// Start entry point for providing config to do the migration
func Start(config *config.Config, storeName string) error {
	if storeName != "Uploads" && storeName != "Avatars" {
		return errors.New("Invalid Store Name")
	}

	log.Println("Store: ", storeName)

	if config.Database.ConnectionString == "" {
		return errors.New("Missing connectionString for Rocket.Chat's Mongo")
	}

	if config.Database.Database == "" {
		return errors.New("Missing db for Rocket.Chat's DB")
	}

	log.Println(config.Source.GoogleStorage.Bucket)

	switch config.Source.Type {
	case "GridFS":
		session, err := connectDB(config.Database.ConnectionString)
		if err != nil {
			return err
		}

		sourceStore = &fileStores.GridFS{
			Database: config.Database.Database,
			Session:  session,
		}
	case "GoogleStorage":
		if config.Source.GoogleStorage.Bucket == "" || config.Source.GoogleStorage.JSONKey == "" {
			return errors.New("Make sure you include all of the required options for GoogleStorage")
		}

		sourceStore = &fileStores.GoogleStorage{
			JSONKey: config.Source.GoogleStorage.JSONKey,
			Bucket:  config.Source.GoogleStorage.Bucket,
		}
	case "AmazonS3":
		if config.Source.AmazonS3.AccessID == "" || config.Source.AmazonS3.AccessKey == "" || config.Source.AmazonS3.Bucket == "" {
			return errors.New("Make sure you include all of the required options for AmazonS3")
		}

		sourceStore = &fileStores.S3{
			Endpoint:  config.Source.AmazonS3.Endpoint,
			AccessId:  config.Source.AmazonS3.AccessID,
			AccessKey: config.Source.AmazonS3.AccessKey,
			Region:    config.Source.AmazonS3.Region,
			Bucket:    config.Source.AmazonS3.Bucket,
			UseSSL:    config.Source.AmazonS3.UseSSL,
		}
	case "FileSystem":
		if config.Source.FileSystem.Location == "" {
			return errors.New("Make sure you include all of the required options for FileSystem")
		}

		sourceStore = &fileStores.FileSystem{
			Location: config.Source.FileSystem.Location,
		}
	default:
		return errors.New("Invalid Source Type")
	}

	log.Println("Source store type set to: ", config.Source.Type)

	switch config.Destination.Type {
	case "AmazonS3":
		if config.Destination.AmazonS3.AccessID == "" || config.Destination.AmazonS3.AccessKey == "" || config.Destination.AmazonS3.Bucket == "" {
			return errors.New("Make sure you include all of the required options for AmazonS3")
		}

		destinationStore = &fileStores.S3{
			Endpoint:  config.Destination.AmazonS3.Endpoint,
			AccessId:  config.Destination.AmazonS3.AccessID,
			AccessKey: config.Destination.AmazonS3.AccessKey,
			Region:    config.Destination.AmazonS3.Region,
			Bucket:    config.Destination.AmazonS3.Bucket,
			UseSSL:    config.Destination.AmazonS3.UseSSL,
		}

	case "GoogleStorage":
		if config.Destination.GoogleStorage.Bucket == "" || config.Destination.GoogleStorage.JSONKey == "" {
			return errors.New("Make sure you include all of the required options for AmazonS3")
		}

		destinationStore = &fileStores.GoogleStorage{
			JSONKey: config.Destination.GoogleStorage.JSONKey,
			Bucket:  config.Destination.GoogleStorage.Bucket,
		}
	case "FileSystem":
		if config.Destination.FileSystem.Location == "" {
			return errors.New("Make sure you include all of the required options for FileSystem")
		}

		sourceStore = &fileStores.FileSystem{
			Location: config.Destination.FileSystem.Location,
		}
	default:
		return errors.New("Invalid Destination Type")
	}

	log.Println("Destination store type set to: ", config.Destination.Type)

	if sourceStore == nil || destinationStore == nil {
		return errors.New("Either source or destination store not initialized")
	}

	err := Migrate(config.Database.ConnectionString, config.Database.Database, sourceStore, destinationStore, storeName)
	if err != nil {
		return err
	}

	return nil
}

func connectDB(connectionstring string) (*mgo.Session, error) {
	log.Println("Connecting to mongodb")

	sess, err := mgo.Dial(connectionstring)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to mongodb")

	return sess.Copy(), nil
}
