package storage

import (
	"context"
	"io"

	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/util"

	"github.com/go-pg/pg/orm"
)

var dataStorage DataStorage

// DataStorage defines common interface for aws and gcloud storage.
type DataStorage interface {
	SafeUpload(ctx context.Context, db orm.DB, bucket, key string, f io.Reader) error
	Upload(ctx context.Context, bucket, key string, f io.Reader) error
	Delete(ctx context.Context, bucket, key string) error
	Client() interface{}
}

// Data returns default initialized global DataStorage.
func Data() DataStorage {
	return dataStorage
}

// InitData sets up a data storage.
func InitData() {
	var err error

	switch settings.Storage.Type {
	case "s3":
		dataStorage = newS3Storage()
	case "gcloud":
		dataStorage, err = newGCloudStorage()
	}

	util.Must(err)
}
