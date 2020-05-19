package storage

import (
	"context"
	"io"

	"github.com/Syncano/orion/pkg/util"

	"github.com/go-pg/pg/v9/orm"
)

type BucketKey string

// DataStorage defines common interface for aws and gcloud storage.
type DataStorage interface {
	Upload(ctx context.Context, bucket BucketKey, key string, f io.Reader) error
	Delete(ctx context.Context, bucket BucketKey, key string) error
	URL(bucket BucketKey, key string) string
}

type bucketInfo struct {
	Name string
	URL  string
}

func SafeUpload(ctx context.Context, storage DataStorage, d *Database, db orm.DB, bucket BucketKey, key string, f io.Reader) error {
	d.AddDBRollbackHook(db, func() error {
		return storage.Delete(ctx, bucket, key)
	})

	return storage.Upload(ctx, bucket, key, f)
}

type Storage struct {
	storageCache     map[string]DataStorage
	buckets          map[BucketKey]string
	loc              string
	host, storageURL string
}

func NewStorage(loc string, buckets map[BucketKey]string, host, storageURL string) *Storage {
	return &Storage{
		storageCache: make(map[string]DataStorage),
		buckets:      buckets,
		loc:          loc,
		host:         host,
		storageURL:   storageURL,
	}
}

func (s *Storage) Default() DataStorage {
	return s.Get(s.loc)
}

func (s *Storage) Get(loc string) DataStorage {
	if s, ok := s.storageCache[loc]; ok {
		return s
	}

	var (
		err         error
		dataStorage DataStorage
	)

	buckets := make(map[BucketKey]*bucketInfo, len(s.buckets))
	for k, v := range s.buckets {
		buckets[k] = &bucketInfo{
			Name: util.GetPrefixEnv(loc, v),
			URL:  util.GetPrefixEnv(loc, v+"_URL"),
		}
	}

	switch util.GetPrefixEnvDefault(loc, "STORAGE", "local") {
	case "s3":
		dataStorage = newS3Storage(loc, buckets)
	case "local":
		dataStorage = newLocalStorage(loc, buckets, s.host, s.storageURL)
	case "gcs":
		dataStorage, err = newGCloudStorage(loc, buckets)
	default:
		panic("unknown storage type")
	}

	util.Must(err)

	s.storageCache[loc] = dataStorage

	return dataStorage
}
