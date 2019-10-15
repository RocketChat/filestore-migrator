package main

import (
	"flag"

	"github.com/RocketChat/MigrateFileStore"
	"github.com/RocketChat/MigrateFileStore/config"
)

func main() {

	configFile := flag.String("configFile", "config.yaml", "Config File full path. Defaults to current folder")
	storeName := flag.String("storeName", "Uploads", "Store Name.  Options: (Uploads, Avatars)")
	skipErrors := flag.Bool("skipErrors", false, "Skip on error")

	flag.Parse()

	config, err := config.Load(*configFile)
	if err != nil {
		panic(err)
	}

	migrate, err := MigrateFileStore.New(config, skipErrors)
	if err != nil {
		panic(err)
	}

	if err := migrate.MigrateStore(*storeName); err != nil {
		panic(err)
	}

}
