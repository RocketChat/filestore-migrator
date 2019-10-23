package main

import (
	"flag"

	migratefiles "github.com/RocketChat/MigrateFileStore"
	"github.com/RocketChat/MigrateFileStore/config"
)

func main() {

	configFile := flag.String("configFile", "config.yaml", "Config File full path. Defaults to current folder")
	storeName := flag.String("storeName", "Uploads", "Store Name.  Options: (Uploads, Avatars)")
	method := flag.String("method", "migrate", "Migrate/Upload/Download")
	skipErrors := flag.Bool("skipErrors", false, "Skip on error")

	flag.Parse()

	config, err := config.Load(*configFile)
	if err != nil {
		panic(err)
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
		if err := migrate.MigrateStore(); err != nil {
			panic(err)
		}
	case "upload":
		if err := migrate.UploadAll(config.TempFileLocation); err != nil {
			panic(err)
		}
	case "download":
		if err := migrate.DownloadAll(); err != nil {
			panic(err)
		}
	default:
		flag.Usage()
		return
	}

}
