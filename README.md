# MigrateFileStore

Right now the only supported are from gridfs to s3

```
go run main.go -storeName {Uploads|Avatars} -source {GridFS} -db {db} -destination {AmazonS3} -s3Bucket {bucket} -s3Endpoint {s3 endpoint} -s3AccessId {s3 AccessId} -s3AccessKey {s3 accessKey}
```

Optionally you can also specify s3 region via: `-s3Region`