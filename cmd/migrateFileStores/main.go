package cmd

import (
	"flag"

	"github.com/RocketChat/MigrateFileStore"
	"github.com/RocketChat/MigrateFileStore/config"
)

func main() {

	configFile := flag.String("configFile", "config.yaml", "Config File full path. Defaults to current folder")
	storeName := flag.String("storeName", "Uploads", "Store Name.  Options: (Uploads, Avatars)")

	flag.Parse()

	config, err := config.Load(*configFile)
	if err != nil {
		panic(err)
	}

	if err := MigrateFileStore.Start(config, *storeName); err != nil {
		panic(err)
	}

}
