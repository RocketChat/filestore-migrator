package migratefilestore

import (
	"crypto/tls"
	"errors"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/RocketChat/MigrateFileStore/pkg/migratestore/config"
	"github.com/RocketChat/MigrateFileStore/pkg/migratestore/store"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Migrate needs to be initialized to begin any migration
type Migrate struct {
	storeName          string
	skipErrors         bool
	sourceStore        store.Provider
	destinationStore   store.Provider
	databaseName       string
	connectionString   string
	fileCollectionName string
	fileOffset         time.Time
	session            *mgo.Session
	uniqueID           string
	tempFileLocation   string
	fileDelay          time.Duration
	debug              bool
}

// New takes the config and returns an initialized Migrate ready to begin migrations
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

	config.TempFileLocation = strings.TrimSuffix(config.TempFileLocation, "/")

	fileDelay := time.Millisecond * 10

	if config.FileDelay != "" {
		delay, err := time.ParseDuration(config.FileDelay)
		if err != nil {
			log.Println("invalid fileDelay value")
			return nil, err
		}

		fileDelay = delay
	}

	migrate := &Migrate{
		skipErrors:       skipErrors,
		databaseName:     config.Database.Database,
		connectionString: config.Database.ConnectionString,
		tempFileLocation: config.TempFileLocation,
		fileDelay:        fileDelay,
		debug:            config.DebugMode,
	}

	if _, err := os.Stat(config.TempFileLocation + "/uploads"); os.IsNotExist(err) {
		if err := os.MkdirAll(config.TempFileLocation+"/uploads", 0777); err != nil {
			migrate.debugLog(err)
			return nil, errors.New("Temp Directory doesn't exist and unable to create it")
		}
	}

	if _, err := os.Stat(config.TempFileLocation + "/avatars"); os.IsNotExist(err) {
		if err := os.MkdirAll(config.TempFileLocation+"/avatars", 0777); err != nil {
			migrate.debugLog(err)
			return nil, errors.New("Temp Directory doesn't exist and unable to create it")
		}
	}

	if config.Source.Type != "" {

		switch config.Source.Type {
		case "GridFS":
			session, err := connectDB(config.Database.ConnectionString)
			if err != nil {
				return nil, err
			}

			sourceStore := &store.GridFSProvider{
				Database:         config.Database.Database,
				Session:          session,
				TempFileLocation: config.TempFileLocation,
			}

			migrate.sourceStore = sourceStore

		case "GoogleStorage":
			if (config.Source.GoogleStorage.Bucket == "" || config.Source.GoogleStorage.JSONKey == "") && !config.Source.ReferenceOnly {
				return nil, errors.New("Make sure you include all of the required options for GoogleStorage")
			}

			sourceStore := &store.GoogleStorageProvider{
				JSONKey:          config.Source.GoogleStorage.JSONKey,
				Bucket:           config.Source.GoogleStorage.Bucket,
				TempFileLocation: config.TempFileLocation,
			}

			migrate.sourceStore = sourceStore
		case "AmazonS3":
			if (config.Source.AmazonS3.AccessID == "" || config.Source.AmazonS3.AccessKey == "" || config.Source.AmazonS3.Bucket == "") && !config.Source.ReferenceOnly {
				return nil, errors.New("Make sure you include all of the required options for AmazonS3")
			}

			sourceStore := &store.S3Provider{
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
			if config.Source.FileSystem.Location == "" && !config.Source.ReferenceOnly {
				return nil, errors.New("Make sure you include all of the required options for FileSystem")
			}

			config.Source.FileSystem.Location = strings.TrimSuffix(config.Source.FileSystem.Location, "/")

			if _, err := os.Stat(config.Source.FileSystem.Location); os.IsNotExist(err) {
				return nil, errors.New("Filesystem source location does not exist or is unaccessible")
			}

			sourceStore := &store.FileSystemStorageProvider{
				Location:         config.Source.FileSystem.Location,
				TempFileLocation: config.TempFileLocation,
			}

			migrate.sourceStore = sourceStore
		default:
			return nil, errors.New("Invalid Source Type")
		}

		migrate.debugLog("Source store type set to: ", config.Source.Type)
	}

	if config.Destination.Type != "" {

		switch config.Destination.Type {
		case "AmazonS3":
			if config.Destination.AmazonS3.AccessID == "" || config.Destination.AmazonS3.AccessKey == "" || config.Destination.AmazonS3.Bucket == "" {
				return nil, errors.New("Make sure you include all of the required options for AmazonS3")
			}

			destinationStore := &store.S3Provider{
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

			destinationStore := &store.GoogleStorageProvider{
				JSONKey: config.Destination.GoogleStorage.JSONKey,
				Bucket:  config.Destination.GoogleStorage.Bucket,
			}

			migrate.destinationStore = destinationStore
		case "FileSystem":
			if config.Destination.FileSystem.Location == "" {
				return nil, errors.New("Make sure you include all of the required options for FileSystem")
			}

			if _, err := os.Stat(config.Destination.FileSystem.Location); os.IsNotExist(err) {
				if err := os.MkdirAll(config.Destination.FileSystem.Location, 0777); err != nil {
					migrate.debugLog(err)
					return nil, errors.New("filesystem directory doesn't exist and unable to create it")
				}
			}

			destinationStore := &store.FileSystemStorageProvider{
				Location: config.Destination.FileSystem.Location,
			}

			migrate.destinationStore = destinationStore
		default:
			return nil, errors.New("Invalid Destination Type")
		}

		migrate.debugLog("Destination store type set to: ", config.Destination.Type)

	}

	if migrate.sourceStore == nil && migrate.destinationStore == nil {
		return nil, errors.New("At least a source or destination store must be provided")
	}

	return migrate, nil
}

// GetRocketChatStore uses database to build source Store from settings
func GetRocketChatStore(dbConfig config.DatabaseConfig) (*config.MigrateTarget, error) {
	session, err := connectDB(dbConfig.ConnectionString)
	if err != nil {
		return nil, err
	}

	sess := session.Copy()
	defer sess.Close()

	db := sess.DB(dbConfig.Database)

	settingsCollection := db.C("rocketchat_settings")

	var fileUploadStorageType rocketChatSetting

	if err := settingsCollection.Find(bson.M{"_id": "FileUpload_Storage_Type"}).One(&fileUploadStorageType); err != nil {
		return nil, err
	}

	sourceStore := &config.MigrateTarget{}

	switch fileUploadStorageType.Value {
	case "GridFS":
		sourceStore.Type = "GridFS"
		return sourceStore, nil
	case "AmazonS3":
		sourceStore.Type = "AmazonS3"
		var awsAccessID rocketChatSetting

		if err := settingsCollection.Find(bson.M{"_id": "FileUpload_S3_AWSAccessKeyId"}).One(&awsAccessID); err != nil {
			return nil, err
		}

		var awsSecret rocketChatSetting

		if err := settingsCollection.Find(bson.M{"_id": "FileUpload_S3_AWSSecretAccessKey"}).One(&awsSecret); err != nil {
			return nil, err
		}

		var bucket rocketChatSetting

		if err := settingsCollection.Find(bson.M{"_id": "FileUpload_S3_Bucket"}).One(&bucket); err != nil {
			return nil, err
		}

		var region rocketChatSetting

		if err := settingsCollection.Find(bson.M{"_id": "FileUpload_S3_Region"}).One(&region); err != nil {
			return nil, err
		}

		var s3url rocketChatSetting

		if err := settingsCollection.Find(bson.M{"_id": "FileUpload_S3_BucketURL"}).One(&s3url); err != nil {
			return nil, err
		}

		if s3url.Value == "" {
			s3url.Value = "s3.amazonaws.com"
		}

		sourceStore.AmazonS3.Endpoint = s3url.Value
		sourceStore.AmazonS3.Bucket = bucket.Value
		sourceStore.AmazonS3.AccessID = awsAccessID.Value
		sourceStore.AmazonS3.AccessKey = awsSecret.Value
		sourceStore.AmazonS3.Region = region.Value
		sourceStore.AmazonS3.UseSSL = true

		return sourceStore, nil

	/*case "GoogleCloudStorage":
	sourceStore.Type = "GoogleStorage"*/

	case "FileSystem":
		sourceStore.Type = "FileSystem"
		var filesystemLocation rocketChatSetting

		if err := settingsCollection.Find(bson.M{"_id": "FileUpload_FileSystemPath"}).One(&filesystemLocation); err != nil {
			return nil, err
		}

		sourceStore.FileSystem.Location = filesystemLocation.Value
		return sourceStore, nil

	default:
		return nil, errors.New("unable to detect supported fileupload storage type.  (Unable to detect google storage currently)")
	}

	return nil, nil
}

func connectDB(connectionstring string) (*mgo.Session, error) {

	ssl := false
	secondaryPreferred := false

	if strings.Contains(connectionstring, "ssl=true") {
		connectionstring = strings.Replace(connectionstring, "&ssl=true", "", -1)
		connectionstring = strings.Replace(connectionstring, "?ssl=true&", "?", -1)
		ssl = true
	}

	if strings.Contains(connectionstring, "readPreference=secondary") {
		connectionstring = strings.Replace(connectionstring, "&readPreference=secondary", "", -1)
		connectionstring = strings.Replace(connectionstring, "?readPreference=secondary", "", -1)
		secondaryPreferred = true
	}

	dialInfo, err := mgo.ParseURL(connectionstring)
	if err != nil {
		return nil, err
	}

	if ssl {
		tlsConfig := &tls.Config{}
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
			return conn, err
		}
	}

	sess, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return nil, err
	}

	if secondaryPreferred {
		sess.SetMode(mgo.SecondaryPreferred, true)
	}

	return sess.Copy(), nil
}
