package settings

type BucketKey string

const (
	BucketData    = BucketKey("data")
	BucketHosting = BucketKey("hosting")
)

var Buckets = map[BucketKey]string{
	BucketData:    "STORAGE_BUCKET",
	BucketHosting: "STORAGE_HOSTING_BUCKET",
}
