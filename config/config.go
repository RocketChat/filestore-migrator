package config

import (
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

var Config *config

type config struct {
	Database    databaseConfig `yaml:"database"`
	Source      MigrateTarget  `yaml:"source"`
	Destination MigrateTarget  `yaml:"destination"`
}

type databaseConfig struct {
	ConnectionString string `yaml:"connectionString"`
	Database         string `yaml:"database"`
}

type MigrateTarget struct {
	Type          string `yaml:"type"`
	GoogleStorage struct {
		JSONKey string `yaml:"jsonKey"`
		Bucket  string `yaml:"bucket"`
	} `yaml:"GoogleStorage"`
	AmazonS3 struct {
		Endpoint  string `yaml:"endpoint"`
		Bucket    string `yaml:"bucket"`
		AccessID  string `yaml:"accessId"`
		AccessKey string `yaml:"accessKey"`
		Region    string `yaml:"region"`
		UseSSL    bool   `yaml:"useSSL"`
	} `yaml:"AmazonS3"`
}

func (c *config) Load(filePath string) error {
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
		return err
	}

	if err := yaml.Unmarshal(yamlFile, c); err != nil {
		log.Fatalf("Unmarshal: %v", err)
		return err
	}

	return nil
}

//Load tries to load the configuration file
func Load(filePath string) error {
	Config = new(config)

	if err := Config.Load(filePath); err != nil {
		return err
	}

	return nil
}
