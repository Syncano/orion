package storage

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/util"
)

type s3Storage struct {
	uploader *s3manager.Uploader
	client   *s3.S3
	buckets  map[settings.BucketKey]string
}

func newS3Storage(loc string, buckets map[settings.BucketKey]string) DataStorage {
	accessKeyID := settings.GetLocationEnv(loc, "S3_ACCESS_KEY_ID")
	secretAccessKey := settings.GetLocationEnv(loc, "S3_SECRET_ACCESS_KEY")
	region := settings.GetLocationEnv(loc, "S3_REGION")
	endpoint := settings.GetLocationEnv(loc, "S3_ENDPOINT")

	sess := createS3Session(accessKeyID, secretAccessKey, region, endpoint)
	client := s3.New(sess)
	uploader := s3manager.NewUploaderWithClient(client)

	return &s3Storage{
		uploader: uploader,
		client:   client,
		buckets:  buckets,
	}
}

// Client returns s3 client.
func (s *s3Storage) Client() interface{} {
	return s.client
}

func createS3Session(accessKeyID, secretAccessKey, region, endpoint string) *session.Session {
	conf := aws.Config{
		Region:   aws.String(region),
		Endpoint: aws.String(endpoint),
	}
	creds := credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")
	sess, err := session.NewSession(conf.WithCredentials(creds))
	util.Must(err)

	return sess
}

func (s *s3Storage) SafeUpload(ctx context.Context, db orm.DB, bucket settings.BucketKey, key string, f io.Reader) error {
	AddDBRollbackHook(db, func() error {
		return s.Delete(ctx, bucket, key)
	})

	return s.Upload(ctx, bucket, key, f)
}

func (s *s3Storage) Upload(ctx context.Context, bucket settings.BucketKey, key string, f io.Reader) error {
	_, err := s.uploader.UploadWithContext(ctx,
		&s3manager.UploadInput{
			Bucket: aws.String(s.buckets[bucket]),
			Key:    aws.String(key),
			ACL:    aws.String("public-read"),
			Body:   f,
		})

	return err
}

func (s *s3Storage) Delete(ctx context.Context, bucket settings.BucketKey, key string) error {
	_, err := s.client.DeleteObjectWithContext(
		ctx,
		&s3.DeleteObjectInput{
			Bucket: aws.String(s.buckets[bucket]),
			Key:    aws.String(key),
		})

	return err
}
