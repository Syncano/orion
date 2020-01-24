package validators

type CacheInvalidateForm struct {
	VersionKey string `form:"version_key" validate:"required"`
	Signature  string `form:"signature" validate:"required"`
}
