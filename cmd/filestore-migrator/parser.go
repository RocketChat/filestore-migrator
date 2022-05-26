package main

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	pkg "github.com/RocketChat/filestore-migrator"
	"github.com/RocketChat/filestore-migrator/config"
)

func parseDatabase(url string) (*config.DatabaseConfig, error) {
	if url == "" {
		err := errors.New("The Rocket.Chat database connection information must be provided")
		return nil, err
	}
	database := config.DatabaseConfig{
		ConnectionString: url,
		Database:         url[strings.LastIndex(url, "/")+1 : len(url)],
	}
	return &database, nil
}

func parseTarget(name string, typ string, connstr string, action string) (*config.MigrateTarget, error) {
	if typ != "" {
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
			target := config.MigrateTarget{
				Type: "AmazonS3",
			}

			if name == "source" && action == "upload" {
				return &target, nil
			}

			if connstr == "" {
				return nil, fmt.Errorf("The %s target information is incomplete", name)
			}

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
			target.AmazonS3 = config.MigrateTargetS3{
				Endpoint:  endpoint,
				Bucket:    bucket,
				AccessID:  accessID,
				AccessKey: accessKey,
				Region:    region,
				UseSSL:    ssl,
			}

			return &target, nil
		case "google":
			target := config.MigrateTarget{
				Type: "GoogleStorage",
			}

			if name == "source" && action == "upload" {
				return &target, nil
			}

			if connstr == "" {
				return nil, fmt.Errorf("The %s target information is incomplete", name)
			}

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

			target.GoogleStorage = config.MigrateTargetGoogleStorage{
				JSONKey: key,
				Bucket:  bucket,
			}

			return &target, nil
		case "filesystem":
			fallthrough
		case "fs":
			if connstr == "" {
				return nil, fmt.Errorf("The %s target information is incomplete", name)
			}

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

// Parse transforms the command arguments into a configuration file.
func Parse(configFile string,
	databaseURL string,
	detectSource bool,
	detectDestination bool,
	sourceType string,
	sourceURL string,
	destinationType string,
	destinationURL string,
	tempLocation string,
	verbose bool,
	action string) (*config.Config, error) {
	if configFile == "" {
		configuration := &config.Config{}
		configuration.DebugMode = verbose
		configuration.TempFileLocation = tempLocation

		database, err := parseDatabase(databaseURL)
		if err != nil {
			panic(err)
		}
		configuration.Database = *database

		if detectSource && detectDestination {
			err := errors.New("Cannot auto detect both source and destination targets. Please, pick one")
			return nil, err
		}

		if !detectSource {
			target, err := parseTarget("source", sourceType, sourceURL, action)
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
			target, err := parseTarget("destination", destinationType, destinationURL, action)
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
