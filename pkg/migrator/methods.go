package migrator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/RocketChat/filestore-migrator/v2/pkg/models"
	"github.com/RocketChat/filestore-migrator/v2/pkg/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type rocketChatSetting struct {
	ID    string `bson:"_id"`
	Value string
}

// DebugMode sets debug mode on
func (m *Migrate) DebugMode() {
	m.debug = true
}

func (m *Migrate) debugLog(v ...interface{}) {
	if m.debug {
		logger(v...)
	}
}

func logger(v ...interface{}) {
	all := append([]interface{}{fmt.Sprintf("[%s]", time.Now().Format("01/02/2006 15:04:05"))}, v...)
	log.Println(all...)
}

// SetFileDelay set the delay between
func (m *Migrate) SetFileDelay(duration time.Duration) {
	m.fileDelay = duration
}

// SetStoreName that will be operating on
func (m *Migrate) SetStoreName(storeName string) error {
	if storeName != "Uploads" && storeName != "Avatars" {
		return errors.New("Invalid Store Name")
	}

	m.storeName = storeName

	if m.sourceStore != nil {
		m.sourceStore.SetTempDirectory(m.tempFileLocation + "/" + strings.ToLower(storeName))
	}

	if m.destinationStore != nil {
		m.destinationStore.SetTempDirectory(m.tempFileLocation + "/" + strings.ToLower(storeName))
	}

	return nil
}

func (m *Migrate) getFiles() ([]models.RocketChatFile, error) {
	if m.storeName == "" {
		return nil, errors.New("no store Name")
	}

	m.debugLog("Store: ", m.storeName)

	fileCollection := ""

	switch m.storeName {
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

	m.session = session

	db := session.Client().Database(m.databaseName)

	settingsCollection := db.Collection("rocketchat_settings")

	var uniqueID rocketChatSetting

	if err := settingsCollection.FindOne(context.TODO(), bson.M{"_id": "uniqueID"}).Decode(&uniqueID); err != nil {
		return nil, err
	}

	m.debugLog("uniqueId", uniqueID)
	m.uniqueID = uniqueID.Value

	collection := db.Collection(fileCollection)

	var files []models.RocketChatFile

	m.debugLog(fileCollection, m.sourceStore.StoreType()+":"+m.storeName)

	query := bson.M{"store": m.sourceStore.StoreType() + ":" + m.storeName}

	if !m.fileOffset.IsZero() {
		query["uploadedAt"] = bson.M{"$gte": m.fileOffset}
	}

	if cursor, err := collection.Find(context.TODO(), query, &options.FindOptions{Sort: bson.D{{"uploadedAt", -1}}}); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("No files found")
		}

		return nil, err
	} else {
		if err = cursor.All(context.TODO(), &files); err != nil {
			return nil, err
		}
	}

	return files, nil
}

// MigrateStore migrates a filestore between source and destination
func (m *Migrate) MigrateStore() error {
	if m.sourceStore == nil || m.destinationStore == nil {
		return errors.New("For MigrateStore both a source and destionation store must be provided")
	}

	files, err := m.getFiles()
	if err != nil {
		return err
	}

	m.debugLog(fmt.Sprintf("Found %v files\n", len(files)))

	for i, file := range files {
		index := i + 1 // for logs

		if m.storeName == "Avatars" && file.Rid != "" {
			// https://github.com/RocketChat/Rocket.Chat/blob/a7823c1b0901c510af5bfa994e7b3f96ee10dd91/apps/meteor/app/file-upload/server/lib/FileUpload.ts#L81-L84
			file.IsRoomAvatar = true
		}

		m.debugLog(fmt.Sprintf("[%v/%v] Downloading %s from: %s\n", index, len(files), file.Name, m.sourceStore.StoreType()))

		if !file.Complete {
			m.debugLog(fmt.Sprintf("[%v/%v] File wasn't completed uploading for %s Skipping\n", index, len(files), file.Name))
			continue
		}

		downloadedPath, err := m.sourceStore.Download(m.fileCollectionName, file)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) || m.skipErrors {
				m.debugLog(fmt.Sprintf("[%v/%v] No corresponding file for %s Skipping\n", index, len(files), file.Name))
				err = nil
				continue
			} else {
				return err
			}
		}

		if file.Rid == "" && m.storeName == "Uploads" {
			file.Rid = "undefined"
		}

		if file.UserID == "" {
			file.UserID = "undefined"
		}

		objectPath := m.getObjectPath(&file)

		m.debugLog(fmt.Sprintf("[%v/%v] Uploading to %s to: %s\n", index, len(files), m.destinationStore.StoreType(), objectPath))

		if err := m.destinationStore.Upload(objectPath, downloadedPath, file.Type); err != nil {
			return err
		}

		set, unset := m.fixFileForUpload(&file, objectPath)

		update := bson.M{
			"$set": set,
		}

		if unset != "" {
			update["$unset"] = bson.M{unset: 1}
		}

		db := m.session.Client().Database(m.databaseName)
		collection := db.Collection(m.fileCollectionName)

		if _, err := collection.UpdateOne(context.TODO(), bson.M{"_id": file.ID}, update); err != nil {
			return err
		}

		m.debugLog(fmt.Sprintf("[%v/%v] Completed Uploading %s\n", index, len(files), file.Name))

		time.Sleep(m.fileDelay)

	}

	m.debugLog("Finished!")

	return nil
}

func (m *Migrate) getObjectPath(file *models.RocketChatFile) string {
	objectPath := ""

	switch m.storeName {
	case "Uploads":
		// https://github.com/RocketChat/Rocket.Chat/blob/a7823c1b0901c510af5bfa994e7b3f96ee10dd91/apps/meteor/app/file-upload/server/lib/FileUpload.ts#L58-L60
		objectPath = fmt.Sprintf("%s/%s/%s/%s/%s", m.uniqueID, strings.ToLower(m.storeName), file.Rid, file.UserID, file.ID)
	case "Avatars":
		var pathSuffix string
		if file.IsRoomAvatar {
			// https://github.com/RocketChat/Rocket.Chat/blob/a7823c1b0901c510af5bfa994e7b3f96ee10dd91/apps/meteor/app/file-upload/server/lib/FileUpload.ts#L81-L84
			pathSuffix = "room-" + file.Rid
		} else {
			pathSuffix = file.UserID
		}

		objectPath = fmt.Sprintf("%s/%s/%s", m.uniqueID, strings.ToLower(m.storeName), pathSuffix)
	}

	// FileSystem just dumps them in the folder based on the ID
	if m.destinationStore.StoreType() == "FileSystem" {
		objectPath = file.ID
	}

	return objectPath
}

func (m *Migrate) fixFileForUpload(file *models.RocketChatFile, objectPath string) (models.FileSetOp, string) {
	// what to unset
	unset := ""

	set := models.FileSetOp{}

	switch m.destinationStore.StoreType() {
	case "AmazonS3":
		set.AmazonS3 = &models.AmazonS3{
			Path: objectPath,
		}

		// Set to empty object so won't be saved back
		unset = "GoogleStorage"

	case "GoogleCloudStorage":
		set.GoogleStorage = &models.GoogleStorage{
			Path: objectPath,
		}

		// Set to empty object so won't be saved back
		unset = "AmazonS3"
	case "FileSystem":
	default:
	}

	ufsPath := fmt.Sprintf("/ufs/%s:%s/%s/%s", m.destinationStore.StoreType(), m.storeName, file.ID, file.Name)

	set.Url = m.siteUrl + ufsPath
	set.Path = ufsPath
	set.Store = m.destinationStore.StoreType() + ":" + m.storeName

	return set, unset
}

// SetFileOffset sets an offset for file upload/downloads
func (m *Migrate) SetFileOffset(offset time.Time) error {
	if offset.IsZero() {
		return errors.New("invalid date")
	}

	m.fileOffset = offset

	return nil
}

// DownloadAll downloads all files from a filestore
func (m *Migrate) DownloadAll() error {
	if m.sourceStore == nil {
		return errors.New("For DownloadAll must have a source store provided")
	}

	files, err := m.getFiles()
	if err != nil {
		return err
	}

	m.debugLog(fmt.Sprintf("Found %v files\n", len(files)))

	for i, file := range files {
		index := i + 1 // for logs

		m.debugLog(fmt.Sprintf("[%v/%v] Downloading %s from: %s\n", index, len(files), file.Name, m.sourceStore.StoreType()))

		if !file.Complete {
			fmt.Printf("[%v/%v] rocketchat.File wasn't completed uploading for %s Skipping\n", index, len(files), file.Name)
			continue
		}

		if _, err := m.sourceStore.Download(m.fileCollectionName, file); err != nil {
			if err == store.ErrNotFound || m.skipErrors {
				fmt.Printf("[%v/%v] No corresponding file for %s Skipping\n", index, len(files), file.Name)
				err = nil
				continue
			} else {
				return err
			}
		}

		m.debugLog(fmt.Sprintf("[%v/%v] Downloaded %s from: %s\n", index, len(files), file.Name, m.sourceStore.StoreType()))

		time.Sleep(m.fileDelay)
	}

	m.debugLog("Finished!")

	return nil
}

// UploadAll uploads all files from a filestore
func (m *Migrate) UploadAll(filesRoot string) error {
	if m.destinationStore == nil {
		return errors.New("For UploadAll must have a destination store provided")
	}

	files, err := m.getFiles()
	if err != nil {
		return err
	}

	m.debugLog(fmt.Sprintf("Found %v files in database\n", len(files)))

	filesRoot = filesRoot + "/" + strings.ToLower(m.storeName)

	for i, file := range files {
		index := i + 1 // for logs

		fileLocation := filesRoot + "/" + file.ID

		if _, err := os.Stat(fileLocation); os.IsNotExist(err) {
			log.Println("Failed to locate: ", file.Name)
			continue
		}

		m.debugLog(fmt.Sprintf("[%v/%v] Uploading %s to: %s\n", index, len(files), file.Name, m.destinationStore.StoreType()))

		if !file.Complete {
			fmt.Printf("[%v/%v] rocketchat.File wasn't completed uploading for %s Skipping\n", index, len(files), file.Name)
			continue
		}

		objectPath := m.getObjectPath(&file)

		m.debugLog(fmt.Sprintf("[%v/%v] Uploading to %s to: %s\n", index, len(files), m.destinationStore.StoreType(), objectPath))
		if err := m.destinationStore.Upload(objectPath, fileLocation, file.Type); err != nil {
			return err
		}

		set, unset := m.fixFileForUpload(&file, objectPath)

		update := bson.M{
			"$set": set,
		}

		if unset != "" {
			update["$unset"] = bson.M{unset: 1}
		}

		collection := m.session.Client().Database(m.databaseName).Collection(m.fileCollectionName)

		if _, err := collection.UpdateOne(context.TODO(), bson.M{"_id": file.ID}, update); err != nil {
			return err
		}

		m.debugLog(fmt.Sprintf("[%v/%v] Completed Uploading %s\n", index, len(files), file.Name))

		time.Sleep(m.fileDelay)
	}

	m.debugLog("Finished!")

	return nil
}
