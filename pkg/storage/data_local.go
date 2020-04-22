package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Syncano/orion/pkg/settings"
)

type localStorage struct {
	basePath string
	buckets  map[settings.BucketKey]*bucketInfo
}

func newLocalStorage(loc string, buckets map[settings.BucketKey]*bucketInfo) DataStorage {
	basePath := settings.GetLocationEnvDefault(loc, "BASE_PATH", "media")

	return &localStorage{
		basePath: basePath,
		buckets:  buckets,
	}
}

func (s *localStorage) URL(bucket settings.BucketKey, key string) string {
	return fmt.Sprintf("http://%s%s%s", settings.API.Host, settings.API.StorageURL, key)
}

func (s *localStorage) Upload(ctx context.Context, bucket settings.BucketKey, key string, f io.Reader) error {
	of, err := os.Create(filepath.Join(s.basePath, key))
	if err != nil {
		return err
	}

	_, err = io.Copy(of, f)

	return err
}

func (s *localStorage) Delete(ctx context.Context, bucket settings.BucketKey, key string) error {
	return os.Remove(filepath.Join(s.basePath, key))
}
