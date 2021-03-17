package config

import (
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

var _config *Config

// Config is the configuration object that will be deserialized from yaml and passed into the migrate client
type Config struct {
	Database         DatabaseConfig `yaml:"database"`
	Source           MigrateTarget  `yaml:"source"`
	Destination      MigrateTarget  `yaml:"destination"`
	TempFileLocation string         `yaml:"tempFileLocation"`
	DebugMode        bool           `yaml:"debugMode"`
	FileDelay        string         `yaml:"fileDelay"`
}

// DatabaseConfig configuration to connect to database
type DatabaseConfig struct {
	ConnectionString string `yaml:"connectionString"`
	Database         string `yaml:"database"`
}

// MigrateTarget is a FileStore configuration for either source or destination
type MigrateTarget struct {
	Type          string                     `yaml:"type"`
	ReferenceOnly bool                       `yaml:"-"`
	GoogleStorage MigrateTargetGoogleStorage `yaml:"GoogleStorage"`
	AmazonS3      MigrateTargetS3            `yaml:"AmazonS3"`
	FileSystem    MigrateTargetFileSystem    `yaml:"FileSystem"`
}

type MigrateTargetGoogleStorage struct {
	JSONKey string `yaml:"jsonKey"`
	Bucket  string `yaml:"bucket"`
}

type MigrateTargetS3 struct {
	Endpoint  string `yaml:"endpoint"`
	Bucket    string `yaml:"bucket"`
	AccessID  string `yaml:"accessId"`
	AccessKey string `yaml:"accessKey"`
	Region    string `yaml:"region"`
	UseSSL    bool   `yaml:"useSSL"`
}

type MigrateTargetFileSystem struct {
	Location string `yaml:"location"`
}

// Get returns the config
func Get() *Config {
	return _config
}

// Load loads the config from file
func (c *Config) Load(filePath string) error {
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
func Load(filePath string) (*Config, error) {
	_config = new(Config)

	if err := _config.Load(filePath); err != nil {
		return nil, err
	}

	return _config, nil
}
