# filestore-migrator

**Provided as-is. Make use of backups and use at your own risk**

**filestore-migrator** is a tool to move files uploaded to a Rocket.Chat instance between object storage providers. Currently we support as targets any object storage provider compatible with the S3 API, as well as, the local file system and Google Cloud Storage. GridFS is also compatible as a source target only.

## Installation

If you are installing **filestore-migrator** from source, make sure that you have a recent version of the `go` runtime installed and that you `$GOPATH` is in your OS system-wide `PATH`. Then, clone the repo to `$GOPATH/src/github.com/RocketChat/filestore-migrator` or use `go get github.com/RocketChat/filestore-migrator`. Now you only need to run the following commands on the tool source directory:

```
make build
filestore-migrator --help
```

## Usage

```
Usage of filestore-migrator:
  -action string
    	Type of action to me performed by the tool (migrate, upload, download ) (default "download")
  -config string
    	Config File full path. Defaults to current folder
  -databaseUrl string
    	Rocket.Chat database connection string
  -destinationType string
    	Destination storage provider (s3, google, fs) (default "s3")
  -destinationUrl string
    	Destination connection string
  -detectDestination
    	Autodetect the destionation using the Rocket.Chat configuration
  -detectSource
    	Autodetect the source target using the Rocket.Chat configuration (default true)
  -skipErrors
    	Skip on error
  -sourceType string
    	Source storage provider (s3, google, gridfs, filesystem) (default "s3")
  -sourceUrl string
    	Source connection string
  -store string
    	Name of the storage to be used in the operation (default "Uploads")
  -tempLocation string
    	Temporary file location (default "/tmp/filestore-migrator")
  -verbose
    	Enable verbose logs (default true)
```

**filestore-migrator** accepts parameters either via flags or via a yaml configuration file, which is examplified in the `cmd` directory. Be aware that each URL type flag have specific patterns, as shown below:

- `databaseUrl`: Rocket.Chat database connection string. Use the official supported mongo connection string-
- `sourceUrl`: Source storage provider (s3, google, gridfs, filesystem)
    - **gridfs**: Automatically retrieved from the Rocket.Chat instance database
    - **s3**: `http://${endpoint}/${bucket_name}?ssl=${ssl}&region=${region}&accessId=${accessId}&accessKey=${accessKey}`
    - **google**: `${json_key}/${bucket_name}`
    - **filesystem**: Normal OS path
- `destinationUrl`: Destination storage provider (s3, google, fs)
    - **s3**: `http://${endpoint}/${bucket_name}?ssl=${ssl}&region=${region}&accessId=${accessId}&accessKey=${accessKey}`
    - **google**: `${json_key}/${bucket_name}`
    - **filesystem**: Normal OS path

## Running with Docker

For those who prefer using **filestore-migrator** via docker, we provide a `Dockerfile` on the root of the directory. First you will need to
compile the image using `make docker`. After the build process, execute with your parameters:

```
docker run filestore-migrator:latest -databaseUrl=mongodb://mongo:27017/rocketchat ...
```
