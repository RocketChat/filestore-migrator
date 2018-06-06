# MigrateFileStore

Right now the only supported are from gridfs to s3

## Installation
If you have go installed: `go get github.com/RocketChat/MigrateFileStore`

Make sure you have $GOPATH/bin in your PATH and then you can run:

```
MigrateFileStore -storeName {Uploads|Avatars} -source {GridFS} -db {db} -destination {AmazonS3} -s3Bucket {bucket} -s3Endpoint {s3 endpoint} -s3AccessId {s3 AccessId} -s3AccessKey {s3 accessKey}
```

Optionally you can also specify s3 region via: `-s3Region`
