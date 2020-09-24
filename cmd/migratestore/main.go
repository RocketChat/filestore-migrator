package main

import (
	"flag"
	"log"

	pkg "github.com/RocketChat/MigrateFileStore/pkg/migratestore"
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
	tempLocation := flag.String("tempLocation", "/tmp/rocket.chat/cloud/migratestore", "Temporary file location")
	store := flag.String("store", "Uploads", "Name of the storage to be used in the operation")
	action := flag.String("action", "download", "Type of action to me performed by the tool (migrate, upload, download )")
	skipErrors := flag.Bool("skipErrors", false, "Skip on error")
	verbose := flag.Bool("verbose", true, "Enable verbose logs")

	flag.Parse()

	config, err := Parse(*configFile,
		*databaseURL,
		*detectSource,
		*detectDestination,
		*sourceType,
		*sourceURL,
		*destinationType,
		*destinationURL,
		*tempLocation,
		*verbose)
	if err != nil {
		panic(err)
	}

	migrate, err := pkg.New(config, *skipErrors)
	if err != nil {
		panic(err)
	}

	if err := migrate.SetStoreName(*store); err != nil {
		panic(err)
	}

	switch *action {
	case "migrate":
		log.Println("Beginning migration of files")
		if err := migrate.MigrateStore(); err != nil {
			panic(err)
		}
	case "upload":
		log.Println("Beginning upload of files")
		if err := migrate.UploadAll(config.TempFileLocation); err != nil {
			panic(err)
		}
	case "download":
		log.Println("Beginning download of files")
		if err := migrate.DownloadAll(); err != nil {
			panic(err)
		}
	default:
		flag.Usage()
		return
	}

	log.Println("Finished!")
}
