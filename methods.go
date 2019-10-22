package MigrateFileStore

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RocketChat/MigrateFileStore/models"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type uniqueIDSetting struct {
	ID    string `bson:"_id"`
	Value string
}

func (m *Migrate) getFiles(storeName string) ([]models.File, error) {
	if storeName != "Uploads" && storeName != "Avatars" {
		return nil, errors.New("Invalid Store Name")
	}

	log.Println("Store: ", storeName)

	fileCollection := ""

	switch storeName {
	case "Uploads":
		fileCollection = "rocketchat_uploads"
	case "Avatars":
		fileCollection = "rocketchat_avatars"
	default:
		return nil, errors.New("Invalid store Name")
	}

	m.fileCollectionName = fileCollection

	session, err := connectDB(m.connectionString)
	if err != nil {
		return nil, err
	}

	sess := session.Copy()
	defer sess.Close()

	db := sess.DB(m.databaseName)

	settingsCollection := db.C("rocketchat_settings")

	var uniqueID uniqueIDSetting

	err = settingsCollection.Find(bson.M{"_id": "uniqueID"}).One(&uniqueID)
	if err != nil {
		return nil, err
	}

	log.Println("uniqueId", uniqueID)
	m.uniqueID = uniqueID.Value

	collection := db.C(fileCollection)

	m.fileCollection = collection

	var files []models.File

	log.Println(fileCollection, m.sourceStore.StoreType()+":"+storeName)

	err = collection.Find(bson.M{"store": m.sourceStore.StoreType() + ":" + storeName}).All(&files)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.New("No files found")
		}

		return nil, err
	}

	return files, nil
}

// MigrateStore migrates a filestore between source and destination
func (m *Migrate) MigrateStore(storeName string) error {
	if m.sourceStore == nil || m.destinationStore == nil {
		return errors.New("For MigrateStore both a source and destionation store must be provided")
	}

	files, err := m.getFiles(storeName)
	if err != nil {
		return err
	}

	fmt.Printf("Found %v files\n", len(files))

	for i, file := range files {
		index := i + 1 // for logs

		fmt.Printf("[%v/%v] Downloading %s from: %s\n", index, len(files), file.Name, m.sourceStore.StoreType())

		if !file.Complete {
			fmt.Printf("[%v/%v] File wasn't completed uploading for %s Skipping\n", index, len(files), file.Name)
			continue
		}

		downloadedPath, err := m.sourceStore.Download(m.fileCollectionName, file)
		if err != nil {
			if err == models.ErrNotFound || m.skipErrors {
				fmt.Printf("[%v/%v] No corresponding file for %s Skipping\n", index, len(files), file.Name)
				err = nil
				continue
			} else {
				return err
			}
		}

		if file.Rid == "" && storeName == "Uploads" {
			file.Rid = "undefined"
		}

		if file.UserId == "" {
			file.UserId = "undefined"
		}

		objectPath := m.getObjectPath(&file, storeName)

		fmt.Printf("[%v/%v] Uploading to %s to: %s\n", index, len(files), m.destinationStore.StoreType(), objectPath)
		err = m.destinationStore.Upload(objectPath, downloadedPath, file.Type)
		if err != nil {
			return err
		}

		unset := m.fixFileForUpload(&file, objectPath, storeName)

		update := bson.M{
			"$set": file,
		}

		if unset != "" {
			update["$unset"] = bson.M{unset: 1}
		}

		if err := m.fileCollection.Update(bson.M{"_id": file.ID}, update); err != nil {
			return err
		}

		fmt.Printf("[%v/%v] Completed Uploading %s\n", index, len(files), file.Name)

		time.Sleep(time.Second * 1)
	}

	log.Println("Finished!")

	return nil
}

func (m *Migrate) getObjectPath(file *models.File, storeName string) string {
	objectPath := ""

	switch storeName {
	case "Uploads":
		objectPath = fmt.Sprintf("%s/%s/%s/%s/%s", m.uniqueID, strings.ToLower(storeName), file.Rid, file.UserId, file.ID)
	case "Avatars":
		objectPath = fmt.Sprintf("%s/%s/%s", m.uniqueID, strings.ToLower(storeName), file.UserId)
	}

	// FileSystem just dumps them in the folder based on the ID
	if m.destinationStore.StoreType() == "FileSystem" {
		objectPath = file.ID
	}

	return objectPath
}

func (m *Migrate) fixFileForUpload(file *models.File, objectPath string, storeName string) string {
	unset := ""

	switch m.destinationStore.StoreType() {
	case "AmazonS3":
		file.AmazonS3 = models.AmazonS3{
			Path: objectPath,
		}

		// Set to empty object so won't be saved back
		unset = "GoogleStorage"
		file.GoogleStorage = models.GoogleStorage{}

	case "GoogleCloudStorage":
		file.GoogleStorage = models.GoogleStorage{
			Path: objectPath,
		}

		// Set to empty object so won't be saved back
		unset = "AmazonS3"
		file.AmazonS3 = models.AmazonS3{}
	case "FileSystem":
	default:
	}

	ufsPath := fmt.Sprintf("/ufs/%s:%s/%s/%s", m.destinationStore.StoreType(), storeName, file.ID, file.Name)

	file.Url = ufsPath
	file.Path = ufsPath
	file.Store = m.destinationStore.StoreType() + ":" + storeName

	return unset
}

// DownloadAll downloads all files from a filestore
func (m *Migrate) DownloadAll(storeName string) error {
	if m.sourceStore == nil {
		return errors.New("For DownloadAll must have a source store provided")
	}

	files, err := m.getFiles(storeName)
	if err != nil {
		return err
	}

	fmt.Printf("Found %v files\n", len(files))

	for i, file := range files {
		index := i + 1 // for logs

		fmt.Printf("[%v/%v] Downloading %s from: %s\n", index, len(files), file.Name, m.sourceStore.StoreType())

		if !file.Complete {
			fmt.Printf("[%v/%v] File wasn't completed uploading for %s Skipping\n", index, len(files), file.Name)
			continue
		}

		_, err := m.sourceStore.Download(m.fileCollectionName, file)
		if err != nil {
			if err == models.ErrNotFound || m.skipErrors {
				fmt.Printf("[%v/%v] No corresponding file for %s Skipping\n", index, len(files), file.Name)
				err = nil
				continue
			} else {
				return err
			}
		}

		fmt.Printf("[%v/%v] Downloaded %s from: %s\n", index, len(files), file.Name, m.sourceStore.StoreType())

		time.Sleep(time.Second * 1)
	}

	log.Println("Finished!")

	return nil
}

// UploadAll uploads all files from a filestore
func (m *Migrate) UploadAll(storeName string, filesRoot string) error {
	if m.destinationStore == nil {
		return errors.New("For UploadAll must have a destination store provided")
	}

	files, err := m.getFiles(storeName)
	if err != nil {
		return err
	}

	fmt.Printf("Found %v files in database\n", len(files))

	for i, file := range files {
		index := i + 1 // for logs

		fileLocation := filesRoot + "/" + file.ID

		if _, err := os.Stat(fileLocation); os.IsNotExist(err) {
			log.Println("Failed to locate: ", file.Name)
			continue
		}

		fmt.Printf("[%v/%v] Uploading %s to: %s\n", index, len(files), file.Name, m.destinationStore.StoreType())

		if !file.Complete {
			fmt.Printf("[%v/%v] File wasn't completed uploading for %s Skipping\n", index, len(files), file.Name)
			continue
		}

		objectPath := m.getObjectPath(&file, storeName)

		fmt.Printf("[%v/%v] Uploading to %s to: %s\n", index, len(files), m.destinationStore.StoreType(), objectPath)
		err = m.destinationStore.Upload(objectPath, fileLocation, file.Type)
		if err != nil {
			return err
		}

		unset := m.fixFileForUpload(&file, objectPath, storeName)

		update := bson.M{
			"$set": file,
		}

		if unset != "" {
			update["$unset"] = bson.M{unset: 1}
		}

		if err := m.fileCollection.Update(bson.M{"_id": file.ID}, update); err != nil {
			return err
		}

		fmt.Printf("[%v/%v] Completed Uploading %s\n", index, len(files), file.Name)

		time.Sleep(time.Second * 1)
	}

	log.Println("Finished!")

	return nil
}
