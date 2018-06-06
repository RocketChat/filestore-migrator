package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	minio "github.com/minio/minio-go"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Upload struct {
	ID          string `bson:"_id"`
	Name        string
	Size        int
	Type        string
	Rid         string
	UserId      string `bson:"userId"`
	Description string
	Store       string
	Complete    bool
	Uploading   bool
	Extension   string
	Progress    int
	AmazonS3    AmazonS3  `bson:"AmazonS3"`
	UpdatedAt   time.Time `bson:"_updatedAt"`
	InstanceID  string    `bson:"instanceId"`
	Etag        string
	Token       string
	UploadedAt  time.Time `bson:"uploadedAt"`
	Path        string
	Url         string
}

type UniqueID struct {
	ID    string `bson:"_id"`
	Value string
}

type AmazonS3 struct {
	Path string
}

type S3Config struct {
	Endpoint  string
	Bucket    string
	AccessId  string
	AccessKey string
	Region    string
	UseSSL    bool
}

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

	if *source == "GridFS" && *dbName == "" {
		log.Println("Missing DB Name for GridFS")
		flag.Usage()
		return
	}

	var s3Config S3Config

	if *destination == "AmazonS3" {
		s3Config = S3Config{
			Endpoint:  *s3_endpoint,
			AccessId:  *s3_accessId,
			AccessKey: *s3_accessKey,
			Region:    *s3_region,
			Bucket:    *s3_bucket,
			UseSSL:    *s3_ssl,
		}

		err := MigrateGridFStoS3(*connectionString, *dbName, s3Config, *storeName)
		if err != nil {
			panic(err)
		}
	}

}

func MigrateGridFStoS3(connectionString string, dbName string, s3Config S3Config, storeName string) error {
	fileCollection := ""

	switch storeName {
	case "Uploads":
		fileCollection = "rocketchat_uploads"
	case "Avatars":
		fileCollection = "rocketchat_avatars"
	default:
		return errors.New("Invalid store Name")
	}

	session, err := ConnectDB(connectionString)
	if err != nil {
		panic(err)
	}

	sess := session.Copy()
	defer sess.Close()

	db := sess.DB(dbName)

	settingsCollection := db.C("rocketchat_settings")

	var uniqueId UniqueID

	err = settingsCollection.Find(bson.M{"_id": "uniqueID"}).One(&uniqueId)
	if err != nil {
		panic(err)
	}

	log.Println("uniqueId", uniqueId)

	collection := db.C(fileCollection)

	var uploads []Upload

	err = collection.Find(bson.M{"store": "GridFS:" + storeName}).All(&uploads)
	if err != nil {
		if err == mgo.ErrNotFound {
			return errors.New("No GridFS files found")
		}

		return err
	}

	fmt.Printf("Found %v files in GridFS\n", len(uploads))

	for i, upload := range uploads {
		fmt.Printf("[%v/%v] Downloading %s\n", i, len(uploads), upload.Name)

		gridFile, err := sess.DB(dbName).GridFS(fileCollection).Open(upload.ID)
		if err != nil {
			if err == mgo.ErrNotFound {
				fmt.Printf("[%v/%v] No corresponding GridFS file for %s Skipping\n", i, len(uploads), upload.Name)
				continue
			}

			return err
		}

		file, err := os.Create("files/" + upload.ID)
		if err != nil {
			return err
		}

		_, err = io.Copy(file, gridFile)
		if err != nil {
			return err
		}

		err = file.Close()
		if err != nil {
			return err
		}

		err = gridFile.Close()
		if err != nil {
			return err
		}

		if upload.Rid == "" && storeName == "Uploads" {
			upload.Rid = "undefined"
		}

		if upload.UserId == "" {
			upload.UserId = "undefined"
		}

		objectPath := ""

		switch storeName {
		case "Uploads":
			objectPath = fmt.Sprintf("%s/%s/%s/%s/%s", uniqueId.Value, strings.ToLower(storeName), upload.Rid, upload.UserId, upload.ID)
		case "Avatars":
			objectPath = fmt.Sprintf("%s/%s/%s", uniqueId.Value, strings.ToLower(storeName), upload.UserId)
		}

		fmt.Printf("[%v/%v] Uploading to S3 to: %s\n", i, len(uploads), objectPath)
		err = UploadFile(s3Config, objectPath, "files/"+upload.ID, upload.Type)
		if err != nil {
			return err
		}

		upload.AmazonS3 = AmazonS3{
			Path: objectPath,
		}

		ufsPath := fmt.Sprintf("/ufs/AmazonS3:%s/%s/%s", storeName, upload.ID, upload.Name)

		upload.Url = ufsPath
		upload.Path = ufsPath
		upload.Store = "AmazonS3:" + storeName

		err = collection.Update(bson.M{
			"_id": upload.ID,
		},
			bson.M{
				"$set": upload,
			})
		if err != nil {
			return err
		}

		fmt.Printf("[%v/%v] Completed Uploading %s\n", i, len(uploads), upload.Name)

		time.Sleep(time.Second * 1)
	}

	log.Println("Finished!")

	return nil
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

func UploadFile(config S3Config, path string, filePath string, contentType string) error {
	minioClient, err := minio.NewWithRegion(config.Endpoint, config.AccessId, config.AccessKey, config.UseSSL, config.Region)
	if err != nil {
		return err
	}

	_, err = minioClient.FPutObject(config.Bucket, path, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
