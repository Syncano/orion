package controllers

import (
	"crypto/hmac"
	"crypto/sha1" // nolint: gosec
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"kkn.fi/base62"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/orion/app/validators"
	"github.com/Syncano/orion/pkg/util"
)

var keyRegex = regexp.MustCompile(`^[a-f0-9]{40}$`)

const (
	apiKeyQuery  = "api_key"
	apiKeyHeader = "X-API-Key"
)

// Auth handles authenticates admin/api key.
func (ctr *Controller) Auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		form := &validators.APIKeyForm{}

		if api.BindAndValidate(c, form) != nil {
			form.APIKey = util.NonEmptyString(c.QueryParam(apiKeyQuery), c.Request().Header.Get(apiKeyHeader))
		}

		if keyRegex.MatchString(form.APIKey) {
			if util.CheckStringParity(form.APIKey) {
				o := &models.Admin{Key: form.APIKey}

				if ctr.q.NewAdminManager(c).OneByKey(o) == nil {
					c.Set(settings.ContextAdminKey, o)
				}
			} else {
				o := &models.APIKey{Key: form.APIKey}

				if ctr.q.NewAPIKeyManager(c).OneByKey(o) == nil {
					c.Set(settings.ContextAPIKeyKey, o)
				}
			}
		} else {
			instance := c.Get(settings.ContextInstanceKey)
			if instance != nil && int64(instance.(*models.Instance).ID) == verifyToken(form.APIKey) {
				o := &models.Admin{ID: instance.(*models.Instance).OwnerID}

				if ctr.q.NewAdminManager(c).OneByID(o) == nil {
					c.Set(settings.ContextAdminKey, o)
				}
			}
		}

		return next(c)
	}
}

func (ctr *Controller) RequireAPIKeyOrAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Get(settings.ContextAPIKeyKey) == nil && c.Get(settings.ContextAdminKey) == nil {
			return api.NewPermissionDeniedError()
		}

		return next(c)
	}
}

func (ctr *Controller) RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Get(settings.ContextAdminKey) == nil {
			return api.NewPermissionDeniedError()
		}

		return next(c)
	}
}

// AuthUser handles authenticates user key.
func (ctr *Controller) AuthUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Get(settings.ContextInstanceKey) != nil {
			form := &validators.UserKeyForm{}
			if api.BindAndValidate(c, form) != nil {
				form.UserKey = util.NonEmptyString(c.QueryParam("user_key"), c.Request().Header.Get("X-User-Key"))
			}

			o := &models.User{Key: form.UserKey}
			if keyRegex.MatchString(form.UserKey) && ctr.q.NewUserManager(c).OneByKey(o) == nil {
				c.Set(settings.ContextUserKey, o)
			}
		}

		return next(c)
	}
}

func (ctr *Controller) RequireUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Get(settings.ContextUserKey) == nil {
			return api.NewPermissionDeniedError()
		}

		return next(c)
	}
}

func createAuthToken(o *models.Instance, expiration time.Duration) string {
	instanceID := base62.Encode(int64(o.ID))
	epoch := base62.Encode(time.Now().Unix() + int64(expiration.Seconds()))
	key := fmt.Sprintf("%s:%s:%s", instanceID, epoch, settings.Common.SecretKey)
	hash := hmac.New(sha1.New, []byte(key)).Sum(nil)

	return fmt.Sprintf("%s:%s:%s", instanceID, epoch, hex.EncodeToString(hash))
}

func verifyToken(token string) int64 {
	t := strings.SplitN(token, ":", 3)
	if len(t) != 3 {
		return -1
	}

	key := fmt.Sprintf("%s:%s:%s", t[0], t[1], settings.Common.SecretKey)
	hash := hmac.New(sha1.New, []byte(key)).Sum(nil)

	if hex.EncodeToString(hash) != t[2] {
		return -1
	}

	if epoch, err := base62.Decode(t[1]); err != nil || epoch < time.Now().Unix() {
		return -1
	}

	i, err := base62.Decode(t[0])
	if err != nil {
		return -1
	}

	return i
}
