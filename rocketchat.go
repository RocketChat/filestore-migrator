package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/RocketChat/MigrateFileStore/fileStores"
	"github.com/RocketChat/MigrateFileStore/models"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type UniqueID struct {
	ID    string `bson:"_id"`
	Value string
}

func Migrate(connectionString string, dbName string, source fileStores.FileStore, destination fileStores.FileStore, storeName string) error {
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

	var files []models.File

	log.Println(fileCollection, source.StoreType()+":"+storeName)

	err = collection.Find(bson.M{"store": source.StoreType() + ":" + storeName}).All(&files)
	if err != nil {
		if err == mgo.ErrNotFound {
			return errors.New("No files found")
		}

		return err
	}

	fmt.Printf("Found %v files\n", len(files))

	for i, file := range files {
		fmt.Printf("[%v/%v] Downloading %s from: %s\n", i, len(files), file.Name, source.StoreType())

		downloadedPath, err := source.Download(fileCollection, file)
		if err != nil {
			if err == models.ErrNotFound {
				fmt.Printf("[%v/%v] No corresponding file for %s Skipping\n", i, len(files), file.Name)
				continue
			} else {
				panic(err)
			}
		}

		if file.Rid == "" && storeName == "Uploads" {
			file.Rid = "undefined"
		}

		if file.UserId == "" {
			file.UserId = "undefined"
		}

		objectPath := ""

		switch storeName {
		case "Uploads":
			objectPath = fmt.Sprintf("%s/%s/%s/%s/%s", uniqueId.Value, strings.ToLower(storeName), file.Rid, file.UserId, file.ID)
		case "Avatars":
			objectPath = fmt.Sprintf("%s/%s/%s", uniqueId.Value, strings.ToLower(storeName), file.UserId)
		}

		fmt.Printf("[%v/%v] Uploading to %s to: %s\n", i, len(files), destination.StoreType(), objectPath)
		err = destination.Upload(objectPath, downloadedPath, file.Type)
		if err != nil {
			return err
		}

		switch destination.StoreType() {
		case "AmazonS3":
			file.AmazonS3 = models.AmazonS3{
				Path: objectPath,
			}
		case "GoogleCloudStorage":
			file.GoogleStorage = models.GoogleStorage{
				Path: objectPath,
			}
		default:
		}

		ufsPath := fmt.Sprintf("/ufs/%s:%s/%s/%s", destination.StoreType(), storeName, file.ID, file.Name)

		file.Url = ufsPath
		file.Path = ufsPath
		file.Store = destination.StoreType() + ":" + storeName

		err = collection.Update(bson.M{
			"_id": file.ID,
		},
			bson.M{
				"$set": file,
			})
		if err != nil {
			return err
		}

		fmt.Printf("[%v/%v] Completed Uploading %s\n", i, len(files), file.Name)

		time.Sleep(time.Second * 1)
	}

	log.Println("Finished!")

	return nil
}
