package storage

import (
	"context"
	"io"

	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/util"

	"github.com/go-pg/pg/v9/orm"
)

// DataStorage defines common interface for aws and gcloud storage.
type DataStorage interface {
	Upload(ctx context.Context, bucket settings.BucketKey, key string, f io.Reader) error
	Delete(ctx context.Context, bucket settings.BucketKey, key string) error
	URL(bucket settings.BucketKey, key string) string
}

type bucketInfo struct {
	Name string
	URL  string
}

func SafeUpload(ctx context.Context, storage DataStorage, db orm.DB, bucket settings.BucketKey, key string, f io.Reader) error {
	AddDBRollbackHook(db, func() error {
		return storage.Delete(ctx, bucket, key)
	})

	return storage.Upload(ctx, bucket, key, f)
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

	switch settings.GetLocationEnvDefault(loc, "STORAGE", "local") {
	case "s3":
		dataStorage = newS3Storage(loc, buckets)
	case "local":
		dataStorage = newLocalStorage(loc, buckets)
	case "gcs":
		dataStorage, err = newGCloudStorage(loc, buckets)
	default:
		panic("unknown storage type")
	}

	util.Must(err)

	storageCache[loc] = dataStorage

	return dataStorage
}
