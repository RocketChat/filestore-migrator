# MigrateFileStore

Supported store types are:
* GridFS
* FileSystem
* AmazonS3
* GoogleStorage

GridFS is not supported as a target you can migrate to.

## Installation
If you have go installed: `go get github.com/RocketChat/MigrateFileStore/cmd/...`

You will need to copy the config.example.yaml to config.yaml and adjust the values

Make sure you have $GOPATH/bin in your PATH and then you can run:

```
migrateFileStores
```

By default it migrates the Uploads store.  If you want to migrate the Avatars store you will need to use the flag:
```
-storeName=Avatars
```

You can also specify the config.yaml path:
```
-configFile=path-to-yaml.yaml
```



