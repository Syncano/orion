package storage

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/Syncano/orion/pkg/util"
)

type gcloudStorage struct {
	client  *storage.Client
	buckets map[BucketKey]*bucketInfo
}

func newGCloudStorage(loc string, buckets map[BucketKey]*bucketInfo) (DataStorage, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(util.GetPrefixEnv(loc, "GOOGLE_APPLICATION_CREDENTIALS")))

	return &gcloudStorage{
		client:  client,
		buckets: buckets,
	}, err
}

// Client returns gcloud client.
func (s *gcloudStorage) Client() interface{} {
	return s.client
}

func (s *gcloudStorage) URL(bucket BucketKey, key string) string {
	return s.buckets[bucket].URL + key
}

func (s *gcloudStorage) Upload(ctx context.Context, bucket BucketKey, key string, f io.Reader) error {
	wc := s.client.Bucket(s.buckets[bucket].Name).Object(key).NewWriter(ctx)
	wc.PredefinedACL = "publicRead"

	if _, err := io.Copy(wc, f); err != nil {
		return err
	}

	return wc.Close()
}

func (s *gcloudStorage) Delete(ctx context.Context, bucket BucketKey, key string) error {
	return s.client.Bucket(s.buckets[bucket].Name).Object(key).Delete(ctx)
}
