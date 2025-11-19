package main

import (
	"flag"
	"log"

	mg "github.com/RocketChat/filestore-migrator/v2/pkg/migrator"
)

func main() {
	configFile := flag.String("config", "", "Config File full path. Defaults to current folder")
	databaseURL := flag.String("databaseUrl", "", "Rocket.Chat database connection string")
	detectSource := flag.Bool("detectSource", true, "Autodetect the source target using the Rocket.Chat configuration")
	detectDestination := flag.Bool("detectDestination", false, "Autodetect the destionation using the Rocket.Chat configuration")
	sourceType := flag.String("sourceType", "s3", "Source storage provider (s3, google, gridfs, filesystem)")
	sourceURL := flag.String("sourceUrl", "", "Source connection string")
	destinationType := flag.String("destinationType", "s3", "Destination storage provider (s3, google, fs)")
	destinationURL := flag.String("destinationUrl", "", "Destination connection string")
	tempLocation := flag.String("tempLocation", "/tmp/filestore-migrator", "Temporary file location")
	store := flag.String("store", "Uploads", "Name of the storage to be used in the operation")
	action := flag.String("action", "download", "Type of action to me performed by the tool (migrate, upload, download )")
	skipErrors := flag.Bool("skipErrors", false, "Skip on error")
	verbose := flag.Bool("verbose", true, "Enable verbose logs")

	flag.Parse()

	// We don't need the source config details. They will have to tell us
	if *action == "upload" {
		*detectSource = false
	}

	if *action == "upload" && *sourceType == "" {
		panic("When specifying upload action you need to provide at least the sourceType")
	}

	config, err := Parse(*configFile,
		*databaseURL,
		*detectSource,
		*detectDestination,
		*sourceType,
		*sourceURL,
		*destinationType,
		*destinationURL,
		*tempLocation,
		*verbose,
		*action)
	if err != nil {
		panic(err)
	}

	migrator, err := mg.New(config, *skipErrors)
	if err != nil {
		panic(err)
	}

	if err := migrator.SetStoreName(*store); err != nil {
		panic(err)
	}

	switch *action {
	case "migrate":
		log.Println("Beginning migration of files")
		if err := migrator.MigrateStore(); err != nil {
			panic(err)
		}
	case "upload":
		log.Println("Beginning upload of files")
		if err := migrator.UploadAll(config.TempFileLocation); err != nil {
			panic(err)
		}
	case "download":
		log.Println("Beginning download of files")
		if err := migrator.DownloadAll(); err != nil {
			panic(err)
		}
	default:
		flag.Usage()
		return
	}

	log.Println("Finished!")
}
