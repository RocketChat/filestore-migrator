package models

import "time"

// RocketChatFile represents the structure of the file in Rocket.Chats database
type RocketChatFile struct {
	ID            string `bson:"_id"`
	Name          string
	Size          int
	Type          string
	Rid           string `bson:"rid"`
	UserID        string `bson:"userId"`
	Description   string
	Store         string
	Complete      bool
	Uploading     bool
	Extension     string
	Progress      float64
	AmazonS3      AmazonS3      `bson:"AmazonS3,omitempty"`
	GoogleStorage GoogleStorage `bson:"GoogleStorage,omitempty"`
	UpdatedAt     time.Time     `bson:"_updatedAt"`
	InstanceID    string        `bson:"instanceId"`
	Identify      struct {
		Format string
		Size   struct {
			Width  int
			Height int
		}
	}
	Etag       string
	Token      string
	UploadedAt time.Time `bson:"uploadedAt"`
	Path       string
	URL        string

	IsRoomAvatar bool
}

type FileSetOp struct {
	GoogleStorage *GoogleStorage `bson:"GoogleStorage,omitempty"`
	AmazonS3      *AmazonS3      `bson:"AmazonS3,omitempty"`
	Url           string         `bson:"url"`
	Path          string         `bson:"path"`
	Store         string         `bson:"store"`
}

// GoogleStorage is sub property of file
type GoogleStorage struct {
	Path string
}

// AmazonS3 is a sub property of file
type AmazonS3 struct {
	Path string
}
