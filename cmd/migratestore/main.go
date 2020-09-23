package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	pkg "github.com/RocketChat/MigrateFileStore/pkg/migratestore"
	"github.com/RocketChat/MigrateFileStore/pkg/migratestore/config"
)

func parseDatabase(url string, name string) (*config.DatabaseConfig, error) {
	if url == "" || name == "" {
		err := errors.New("The Rocket.Chat database connection information must be provided")
		return nil, err
	}
	database := config.DatabaseConfig{
		ConnectionString: url,
		Database:         name,
	}
	return &database, nil
}

func parseTarget(name string, typ string, connstr string) (*config.MigrateTarget, error) {
	if typ != "" && connstr != "" {
		if name == "destination" && typ == "gridfs" {
			err := errors.New("You cannot use gridfs as a destination target")
			return nil, err
		}

		switch typ {
		case "gridfs":
			target := config.MigrateTarget{
				Type: "GridFS",
			}
			return &target, nil
		case "s3":
			urlInfo, err := url.Parse(connstr)
			if err != nil {
				panic(err)
			}
			endpoint := urlInfo.Host
			if endpoint == "" {
				err := errors.New("The informed S3 connection string doesn't contain the endpoint field")
				return nil, err
			}
			bucket := strings.Trim(urlInfo.EscapedPath(), "/")
			if bucket == "" {
				err := errors.New("The informed S3 connection string doesn't contain the bucket field")
				return nil, err
			}
			accessID := urlInfo.Query().Get("accessId")
			if accessID == "" {
				err := errors.New("The informed S3 connection string doesn't contain the access ID field")
				return nil, err
			}
			accessKey := urlInfo.Query().Get("accessKey")
			if accessKey == "" {
				err := errors.New("The informed S3 connection string doesn't contain the access key field")
				return nil, err
			}
			region := urlInfo.Query().Get("region")
			if region == "" {
				err := errors.New("The informed S3 connection string doesn't contain the region field")
				return nil, err
			}
			if urlInfo.Query().Get("ssl") == "" {
				err := errors.New("The informed S3 connection string doesn't contain the ssl field")
				return nil, err
			}
			ssl, err := strconv.ParseBool(urlInfo.Query().Get("ssl"))
			if err != nil {
				panic(err)
			}
			target := config.MigrateTarget{
				Type: "AmazonS3",
				AmazonS3: config.MigrateTargetS3{
					Endpoint:  endpoint,
					Bucket:    bucket,
					AccessID:  accessID,
					AccessKey: accessKey,
					Region:    region,
					UseSSL:    ssl,
				},
			}
			return &target, nil
		case "google":
			info := strings.Split(connstr, "/")
			if len(info) != 2 {
				err := errors.New("The informed Google Cloud connection string doesn't respect the tool pattern")
				return nil, err
			}
			key := info[0]
			if key == "" {
				err := errors.New("The informed Google Cloud connection string doesn't contain the json key field")
				return nil, err
			}
			bucket := info[1]
			if bucket == "" {
				err := errors.New("The informed Google Cloud connection string doesn't contain the bucket field")
				return nil, err
			}
			target := config.MigrateTarget{
				Type: "GoogleStorage",
				GoogleStorage: config.MigrateTargetGoogleStorage{
					JSONKey: key,
					Bucket:  bucket,
				},
			}
			return &target, nil
		case "fs":
			target := config.MigrateTarget{
				Type: "FileSystem",
				FileSystem: config.MigrateTargetFileSystem{
					Location: connstr,
				},
			}
			return &target, nil
		default:
			err := errors.New("The type target informed is not supported")
			return nil, err
		}
	}

	err := fmt.Errorf("The %s target information is incomplete", name)
	return nil, err
}

func parse(configFile string,
	databaseURL string,
	databaseName string,
	detectSource bool,
	detectDestination bool,
	sourceType string,
	sourceURL string,
	destinationType string,
	destinationURL string,
	tempLocation string,
	verbose bool) (*config.Config, error) {
	if configFile == "" {
		configuration := &config.Config{}
		configuration.DebugMode = verbose
		configuration.TempFileLocation = tempLocation

		database, err := parseDatabase(databaseURL, databaseName)
		if err != nil {
			panic(err)
		}
		configuration.Database = *database

		if detectSource && detectDestination {
			err := errors.New("Cannot auto detect both source and destination targets. Please, pick one")
			return nil, err
		}

		if !detectSource {
			target, err := parseTarget("source", sourceType, sourceURL)
			if err != nil {
				panic(err)
			}
			configuration.Source = *target
		} else {
			target, err := pkg.GetRocketChatStore(configuration.Database)
			if err != nil {
				panic(err)
			}
			configuration.Source = *target
		}

		if !detectDestination {
			target, err := parseTarget("destination", destinationType, destinationURL)
			if err != nil {
				panic(err)
			}
			configuration.Destination = *target
		} else {
			target, err := pkg.GetRocketChatStore(configuration.Database)
			if err != nil {
				panic(err)
			}
			configuration.Destination = *target
		}

		return configuration, nil
	}

	configuration, err := config.Load(configFile)
	if err != nil {
		panic(err)
	}

	if verbose {
		configuration.DebugMode = true
	}

	return configuration, nil
}

func main() {
	configFile := flag.String("config", "", "Config File full path. Defaults to current folder")
	databaseURL := flag.String("databaseUrl", "", "Rocket.Chat database connection string")
	databaseName := flag.String("databaseName", "", "Rocket.Chat database name")
	detectSource := flag.Bool("detectSource", true, "Autodetect the source target using the Rocket.Chat configuration")
	detectDestination := flag.Bool("detectDestination", false, "Autodetect the destionation using the Rocket.Chat configuration")
	sourceType := flag.String("sourceType", "s3", "Source storage provider (S3, Google Cloud Storage, GridFS, Filesystem)")
	sourceURL := flag.String("sourceUrl", "", "Source connection string")
	destinationType := flag.String("destinationType", "s3", "Destination storage provider (S3, Google Cloud Storage, GridFS, Filesystem)")
	destinationURL := flag.String("destinationUrl", "", "Destination connection string")
	tempLocation := flag.String("tempLocation", "/tmp/rocket.chat/cloud/migration", "Temporary file location")
	store := flag.String("store", "Uploads", "Name of the storage to be used in the operation (Uploads, Avatars)")
	action := flag.String("action", "download", "Options are: (migrate/upload/download)")
	skipErrors := flag.Bool("skipErrors", false, "Skip on error")
	verbose := flag.Bool("verbose", true, "Enable verbose logs")

	flag.Parse()

	config, err := parse(*configFile,
		*databaseURL,
		*databaseName,
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
