package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Syncano/orion/pkg/util"
)

type localStorage struct {
	basePath         string
	buckets          map[BucketKey]*bucketInfo
	host, storageURL string
}

func newLocalStorage(loc string, buckets map[BucketKey]*bucketInfo, host, storageURL string) DataStorage {
	basePath := util.GetPrefixEnvDefault(loc, "BASE_PATH", "media")

	return &localStorage{
		basePath:   basePath,
		buckets:    buckets,
		host:       host,
		storageURL: storageURL,
	}
}

func (s *localStorage) URL(bucket BucketKey, key string) string {
	return fmt.Sprintf("http://%s%s%s", s.host, s.storageURL, key)
}

func (s *localStorage) Upload(ctx context.Context, bucket BucketKey, key string, f io.Reader) error {
	of, err := os.Create(filepath.Join(s.basePath, key))
	if err != nil {
		return err
	}

	_, err = io.Copy(of, f)

	return err
}

func (s *localStorage) Delete(ctx context.Context, bucket BucketKey, key string) error {
	return os.Remove(filepath.Join(s.basePath, key))
}
