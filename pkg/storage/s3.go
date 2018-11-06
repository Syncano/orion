package storage

import (
	"io"
	"mime/multipart"

	"github.com/Syncano/orion/pkg/settings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/pkg/util"
)

var (
	s3Uploader *s3manager.Uploader

	s3Client *s3.S3
)

// InitS3 sets up common and tenant s3 clients and uploaders.
func InitS3() {
	s3Session := createS3Session(settings.S3.AccessKeyID, settings.S3.SecretAccessKey,
		settings.S3.Region, settings.S3.Endpoint)
	s3Client = s3.New(s3Session)
	s3Uploader = s3manager.NewUploaderWithClient(s3Client)
}

// S3 returns s3 client.
func S3() *s3.S3 {
	return s3Client
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

// SafeUploadFileheaderToS3 ...
func SafeUploadFileheaderToS3(client *s3.S3, db orm.DB, bucket, key string, fh *multipart.FileHeader) error {
	AddDBRollbackHook(db, func() {
		util.Must(
			DeleteS3File(client, bucket, key),
		)
	})
	return UploadFileheaderToS3(client, bucket, key, fh)
}

// UploadFileheaderToS3 ...
func UploadFileheaderToS3(client *s3.S3, bucket, key string, fh *multipart.FileHeader) error {
	f, err := fh.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	return UploadFileToS3(client, bucket, key, f)
}

// SafeUploadFileToS3 ...
func SafeUploadFileToS3(client *s3.S3, db orm.DB, bucket, key string, body io.Reader) error {
	AddDBRollbackHook(db, func() {
		util.Must(
			DeleteS3File(client, bucket, key),
		)
	})
	return UploadFileToS3(client, bucket, key, body)
}

// UploadFileToS3 ...
func UploadFileToS3(client *s3.S3, bucket, key string, body io.Reader) error {
	_, err := s3Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ACL:    aws.String("public-read"),
		Body:   body,
	})
	return err
}

func deleteS3File(client *s3.S3, bucket, key string) error {
	_, err := client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

// DeleteS3File ...
func DeleteS3File(client *s3.S3, bucket, key string) error {
	return deleteS3File(client, bucket, key)
}
