package models

import "time"

type File struct {
	ID          string `bson:"_id"`
	Name        string
	Size        int
	Type        string
	Rid         string
	UserId      string `bson:"userId"`
	Description string
	Store       string
	Complete    bool
	Uploading   bool
	Extension   string
	Progress    int
	AmazonS3    AmazonS3  `bson:"AmazonS3"`
	UpdatedAt   time.Time `bson:"_updatedAt"`
	InstanceID  string    `bson:"instanceId"`
	Etag        string
	Token       string
	UploadedAt  time.Time `bson:"uploadedAt"`
	Path        string
	Url         string
}

type AmazonS3 struct {
	Path string
}
