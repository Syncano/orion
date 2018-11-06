package validators

// APIKeyForm ...
type APIKeyForm struct {
	APIKey string `form:"_api_key" validate:"required"`
}

// UserKeyForm ...
type UserKeyForm struct {
	UserKey string `form:"_user_key" validate:"required"`
}
