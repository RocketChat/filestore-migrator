package MigrateFileStore

import (
	"errors"
	"log"
	"os"

	"github.com/RocketChat/MigrateFileStore/config"
	"github.com/RocketChat/MigrateFileStore/fileStores"
	mgo "gopkg.in/mgo.v2"
)

type Migrate struct {
	storeName          string
	skipErrors         bool
	sourceStore        fileStores.FileStore
	destinationStore   fileStores.FileStore
	databaseName       string
	connectionString   string
	fileCollectionName string
	fileCollection     *mgo.Collection
	uniqueID           string
}

func New(config *config.Config, skipErrors bool) (*Migrate, error) {

	if config.Database.ConnectionString == "" {
		return nil, errors.New("Missing connectionString for Rocket.Chat's Mongo")
	}

	if config.Database.Database == "" {
		return nil, errors.New("Missing db for Rocket.Chat's DB")
	}

	if config.TempFileLocation == "" {
		config.TempFileLocation = "files"
	}

	if _, err := os.Stat(config.TempFileLocation); os.IsNotExist(err) {
		if err := os.MkdirAll(config.TempFileLocation, 0600); err != nil {
			log.Println(err)
			return nil, errors.New("Directory doesn't exist and unable to create it")
		}
	}

	migrate := &Migrate{
		skipErrors:       skipErrors,
		databaseName:     config.Database.Database,
		connectionString: config.Database.ConnectionString,
	}

	if config.Source.Type != "" {

		switch config.Source.Type {
		case "GridFS":
			session, err := connectDB(config.Database.ConnectionString)
			if err != nil {
				return nil, err
			}

			sourceStore := &fileStores.GridFS{
				Database:         config.Database.Database,
				Session:          session,
				TempFileLocation: config.TempFileLocation,
			}

			migrate.sourceStore = sourceStore

		case "GoogleStorage":
			if config.Source.GoogleStorage.Bucket == "" || config.Source.GoogleStorage.JSONKey == "" {
				return nil, errors.New("Make sure you include all of the required options for GoogleStorage")
			}

			sourceStore := &fileStores.GoogleStorage{
				JSONKey:          config.Source.GoogleStorage.JSONKey,
				Bucket:           config.Source.GoogleStorage.Bucket,
				TempFileLocation: config.TempFileLocation,
			}

			migrate.sourceStore = sourceStore
		case "AmazonS3":
			if config.Source.AmazonS3.AccessID == "" || config.Source.AmazonS3.AccessKey == "" || config.Source.AmazonS3.Bucket == "" {
				return nil, errors.New("Make sure you include all of the required options for AmazonS3")
			}

			sourceStore := &fileStores.S3{
				Endpoint:         config.Source.AmazonS3.Endpoint,
				AccessID:         config.Source.AmazonS3.AccessID,
				AccessKey:        config.Source.AmazonS3.AccessKey,
				Region:           config.Source.AmazonS3.Region,
				Bucket:           config.Source.AmazonS3.Bucket,
				UseSSL:           config.Source.AmazonS3.UseSSL,
				TempFileLocation: config.TempFileLocation,
			}

			migrate.sourceStore = sourceStore
		case "FileSystem":
			if config.Source.FileSystem.Location == "" {
				return nil, errors.New("Make sure you include all of the required options for FileSystem")
			}

			sourceStore := &fileStores.FileSystem{
				Location:         config.Source.FileSystem.Location,
				TempFileLocation: config.TempFileLocation,
			}

			migrate.sourceStore = sourceStore
		default:
			return nil, errors.New("Invalid Source Type")
		}

		log.Println("Source store type set to: ", config.Source.Type)
	}

	if config.Destination.Type != "" {

		switch config.Destination.Type {
		case "AmazonS3":
			if config.Destination.AmazonS3.AccessID == "" || config.Destination.AmazonS3.AccessKey == "" || config.Destination.AmazonS3.Bucket == "" {
				return nil, errors.New("Make sure you include all of the required options for AmazonS3")
			}

			destinationStore := &fileStores.S3{
				Endpoint:  config.Destination.AmazonS3.Endpoint,
				AccessID:  config.Destination.AmazonS3.AccessID,
				AccessKey: config.Destination.AmazonS3.AccessKey,
				Region:    config.Destination.AmazonS3.Region,
				Bucket:    config.Destination.AmazonS3.Bucket,
				UseSSL:    config.Destination.AmazonS3.UseSSL,
			}

			migrate.destinationStore = destinationStore

		case "GoogleStorage":
			if config.Destination.GoogleStorage.Bucket == "" || config.Destination.GoogleStorage.JSONKey == "" {
				return nil, errors.New("Make sure you include all of the required options for AmazonS3")
			}

			destinationStore := &fileStores.GoogleStorage{
				JSONKey: config.Destination.GoogleStorage.JSONKey,
				Bucket:  config.Destination.GoogleStorage.Bucket,
			}

			migrate.destinationStore = destinationStore
		case "FileSystem":
			if config.Destination.FileSystem.Location == "" {
				return nil, errors.New("Make sure you include all of the required options for FileSystem")
			}

			destinationStore := &fileStores.FileSystem{
				Location: config.Destination.FileSystem.Location,
			}

			migrate.destinationStore = destinationStore
		default:
			return nil, errors.New("Invalid Destination Type")
		}

		log.Println("Destination store type set to: ", config.Destination.Type)

	}

	if migrate.sourceStore == nil && migrate.destinationStore == nil {
		return nil, errors.New("At least a source or destination store must be provided")
	}

	return migrate, nil
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
