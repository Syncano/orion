package settings

import (
	"github.com/Syncano/pkg-go/v2/storage"
)

const (
	BucketData    = storage.BucketKey("data")
	BucketHosting = storage.BucketKey("hosting")
)

var (
	Buckets = map[storage.BucketKey]string{
		BucketData:    "STORAGE_BUCKET",
		BucketHosting: "STORAGE_HOSTING_BUCKET",
	}
)
