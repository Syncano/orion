package storage

import (
	"context"
	"io"

	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/util"

	"github.com/go-pg/pg/orm"
)

// DataStorage defines common interface for aws and gcloud storage.
type DataStorage interface {
	SafeUpload(ctx context.Context, db orm.DB, bucket settings.BucketKey, key string, f io.Reader) error
	Upload(ctx context.Context, bucket settings.BucketKey, key string, f io.Reader) error
	Delete(ctx context.Context, bucket settings.BucketKey, key string) error
	URL(bucket settings.BucketKey, key string) string
	Client() interface{}
}

type bucketInfo struct {
	Name string
	URL  string
}

var storageCache map[string]DataStorage = make(map[string]DataStorage)

func Default() DataStorage {
	return Storage(settings.Common.Location)
}

func Storage(loc string) DataStorage {
	if s, ok := storageCache[loc]; ok {
		return s
	}

	var (
		err         error
		dataStorage DataStorage
	)

	buckets := make(map[settings.BucketKey]*bucketInfo, len(settings.Buckets))
	for k, v := range settings.Buckets {
		buckets[k] = &bucketInfo{
			Name: settings.GetLocationEnv(loc, v),
			URL:  settings.GetLocationEnv(loc, v+"_URL"),
		}
	}

	switch settings.GetLocationEnv(loc, "STORAGE") {
	case "s3":
		dataStorage = newS3Storage(loc, buckets)
	case "gcs":
		dataStorage, err = newGCloudStorage(loc, buckets)
	default:
		panic("unknown storage type")
	}

	util.Must(err)

	storageCache[loc] = dataStorage

	return dataStorage
}

// InitData sets up a data storage.
func InitData() {
	Default()
}
