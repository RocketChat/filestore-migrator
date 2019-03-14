package main

import (
	"flag"
	"log"

	"github.com/RocketChat/MigrateFileStore/config"
	"github.com/RocketChat/MigrateFileStore/fileStores"
	mgo "gopkg.in/mgo.v2"
)

var sourceStore fileStores.FileStore
var destinationStore fileStores.FileStore

func main() {

	configFile := flag.String("configFile", "config.yaml", "Config File full path. Defaults to current folder")
	storeName := flag.String("storeName", "Uploads", "Store Name.  Options: (Uploads, Avatars)")

	flag.Parse()

	if err := config.Load(*configFile); err != nil {
		panic(err)
	}

	if config.Config.Database.ConnectionString == "" {
		log.Println("Missing connectionString for Rocket.Chat's Mongo")
		flag.Usage()
		return
	}

	if config.Config.Database.Database == "" {
		log.Println("Missing db for Rocket.Chat's DB")
		flag.Usage()
		return
	}

	log.Println(config.Config.Source.GoogleStorage.Bucket)

	switch config.Config.Source.Type {
	case "GridFS":
		session, err := ConnectDB(config.Config.Database.ConnectionString)
		if err != nil {
			panic(err)
		}

		sourceStore = &fileStores.GridFS{
			Database: config.Config.Database.Database,
			Session:  session,
		}
	case "GoogleStorage":
		if config.Config.Source.GoogleStorage.Bucket == "" || config.Config.Source.GoogleStorage.JSONKey == "" {
			log.Println("Make sure you include all of the required options for GoogleStorage")
			return
		}

		sourceStore = &fileStores.GoogleStorage{
			JSONKey: config.Config.Source.GoogleStorage.JSONKey,
			Bucket:  config.Config.Source.GoogleStorage.Bucket,
		}
	case "AmazonS3":
		if config.Config.Source.AmazonS3.AccessID == "" || config.Config.Source.AmazonS3.AccessKey == "" || config.Config.Source.AmazonS3.Bucket == "" {
			log.Println("Make sure you include all of the required options for AmazonS3")
			return
		}

		sourceStore = &fileStores.S3{
			Endpoint:  config.Config.Source.AmazonS3.Endpoint,
			AccessId:  config.Config.Source.AmazonS3.AccessID,
			AccessKey: config.Config.Source.AmazonS3.AccessKey,
			Region:    config.Config.Source.AmazonS3.Region,
			Bucket:    config.Config.Source.AmazonS3.Bucket,
			UseSSL:    config.Config.Source.AmazonS3.UseSSL,
		}
	default:
		log.Println("Invalid Source type")
		return
	}

	log.Println("Source store type set to: ", config.Config.Source.Type)

	switch config.Config.Destination.Type {
	case "AmazonS3":
		if config.Config.Destination.AmazonS3.AccessID == "" || config.Config.Destination.AmazonS3.AccessKey == "" || config.Config.Destination.AmazonS3.Bucket == "" {
			log.Println("Make sure you include all of the required options for AmazonS3")
			return
		}

		destinationStore = &fileStores.S3{
			Endpoint:  config.Config.Destination.AmazonS3.Endpoint,
			AccessId:  config.Config.Destination.AmazonS3.AccessID,
			AccessKey: config.Config.Destination.AmazonS3.AccessKey,
			Region:    config.Config.Destination.AmazonS3.Region,
			Bucket:    config.Config.Destination.AmazonS3.Bucket,
			UseSSL:    config.Config.Destination.AmazonS3.UseSSL,
		}

	case "GoogleStorage":
		if config.Config.Destination.GoogleStorage.Bucket == "" || config.Config.Destination.GoogleStorage.JSONKey == "" {
			log.Println("Make sure you include all of the required options for AmazonS3")
			return
		}

		destinationStore = &fileStores.GoogleStorage{
			JSONKey: config.Config.Destination.GoogleStorage.JSONKey,
			Bucket:  config.Config.Destination.GoogleStorage.Bucket,
		}
	default:
		log.Println("Invalid Destination Type")
		return
	}

	log.Println("Destination store type set to: ", config.Config.Destination.Type)

	if *storeName != "Uploads" && *storeName != "Avatars" {
		log.Println("Invalid Store Name")
		flag.Usage()
		return
	}

	log.Println("Store: ", *storeName)

	if sourceStore == nil || destinationStore == nil {
		panic("Either source or destination store not initialized")
	}

	err := Migrate(config.Config.Database.ConnectionString, config.Config.Database.Database, sourceStore, destinationStore, *storeName)
	if err != nil {
		panic(err)
	}

}

func ConnectDB(connectionstring string) (*mgo.Session, error) {
	log.Println("Connecting to mongodb")

	sess, err := mgo.Dial(connectionstring)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to mongodb")

	return sess.Copy(), nil
}
