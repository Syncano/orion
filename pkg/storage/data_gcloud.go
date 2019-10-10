package storage

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
	"github.com/go-pg/pg/orm"
)

type gcloudStorage struct {
	client *storage.Client
}

func newGCloudStorage() (DataStorage, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)

	return &gcloudStorage{
		client: client,
	}, err
}

// Client returns gcloud client.
func (s *gcloudStorage) Client() interface{} {
	return s.client
}

// SafeUploadFileToS3 ...
func (s *gcloudStorage) SafeUpload(ctx context.Context, db orm.DB, bucket, key string, f io.Reader) error {
	AddDBRollbackHook(db, func() error {
		return s.Delete(ctx, bucket, key)
	})
	return s.Upload(ctx, bucket, key, f)
}

// UploadFileToS3 ...
func (s *gcloudStorage) Upload(ctx context.Context, bucket, key string, f io.Reader) error {
	wc := s.client.Bucket(bucket).Object(key).NewWriter(ctx)
	wc.PredefinedACL = "publicRead"
	if _, err := io.Copy(wc, f); err != nil {
		return err
	}
	return wc.Close()
}

// DeleteS3File ...
func (s *gcloudStorage) Delete(ctx context.Context, bucket, key string) error {
	return s.client.Bucket(bucket).Object(key).Delete(ctx)
}
