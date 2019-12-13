package main

import (
	"flag"
	"log"

	migratefiles "github.com/RocketChat/MigrateFileStore"
	"github.com/RocketChat/MigrateFileStore/config"
)

func main() {

	configFile := flag.String("configFile", "config.yaml", "Config File full path. Defaults to current folder")
	storeName := flag.String("storeName", "Uploads", "Store Name.  Options: (Uploads, Avatars)")
	method := flag.String("method", "download", "Options are: (migrate/upload/download)")
	skipErrors := flag.Bool("skipErrors", false, "Skip on error")
	verbose := flag.Bool("verbose", true, "setting to true will spit out more verbose logs")

	flag.Parse()

	config, err := config.Load(*configFile)
	if err != nil {
		panic(err)
	}

	if *verbose {
		config.DebugMode = true
	}

	migrate, err := migratefiles.New(config, *skipErrors)
	if err != nil {
		panic(err)
	}

	if err := migrate.SetStoreName(*storeName); err != nil {
		panic(err)
	}

	switch *method {
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
