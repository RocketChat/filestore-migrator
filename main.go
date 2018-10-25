package main

import (
	"flag"
	"log"

	"github.com/RocketChat/MigrateFileStore/fileStores"
	mgo "gopkg.in/mgo.v2"
)

var sourceStore fileStores.FileStore
var destinationStore fileStores.FileStore

func main() {

	storeName := flag.String("storeName", "Uploads", "Store Name.  Options: (Uploads, Avatars)")
	source := flag.String("source", "GridFS", "Source of files. Options: (GridFS)")
	destination := flag.String("destination", "AmazonS3", "Destination of files. Options: (AmazonS3)")
	connectionString := flag.String("connectionString", "mongodb://127.0.0.1:27017", "Mongodb connection url")
	dbName := flag.String("db", "", "Mongo database")
	s3_endpoint := flag.String("s3Endpoint", "", "S3 Endpoint")
	s3_region := flag.String("s3Region", "", "S3 region")
	s3_accessId := flag.String("s3AccessId", "", "S3 Access Id")
	s3_accessKey := flag.String("s3AccessKey", "", "S3 Access Key")
	s3_bucket := flag.String("s3Bucket", "", "S3 Bucket")
	s3_ssl := flag.Bool("s3SSL", true, "s3 Use SSL")

	flag.Parse()

	if *connectionString == "" {
		log.Println("Missing connectionString for Rocket.Chat's Mongo")
		flag.Usage()
		return
	}

	if *dbName == "" {
		log.Println("Missing db for Rocket.Chat's DB")
		flag.Usage()
		return
	}

	if *source != "GridFS" {
		log.Println("Invalid Source")
		flag.Usage()
		return
	}

	if *destination != "AmazonS3" {
		log.Println("Invalid Destination")
		flag.Usage()
		return
	}

	if *destination == "AmazonS3" && (*s3_endpoint == "" || *s3_accessId == "" || *s3_accessKey == "" || *s3_bucket == "") {
		log.Println("Please include required options for s3")
		flag.Usage()
		return
	}

	if *storeName != "Uploads" && *storeName != "Avatars" {
		log.Println("Invalid Store Name")
		flag.Usage()
		return
	}

	if *source == "GridFS" {
		session, err := ConnectDB(*connectionString)
		if err != nil {
			panic(err)
		}

		sourceStore = &fileStores.GridFS{
			Database: *dbName,
			Session:  session,
		}
	}

	if *destination == "AmazonS3" {
		destinationStore = &fileStores.S3{
			Endpoint:  *s3_endpoint,
			AccessId:  *s3_accessId,
			AccessKey: *s3_accessKey,
			Region:    *s3_region,
			Bucket:    *s3_bucket,
			UseSSL:    *s3_ssl,
		}
	}

	err := Migrate(*connectionString, *dbName, sourceStore, destinationStore, *storeName)
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
