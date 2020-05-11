package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/orion/app/validators"
)

func (ctr *Controller) CacheInvalidate(c echo.Context) error {
	v := &validators.CacheInvalidateForm{}

	if err := api.BindValidateAndExec(c, v, func() error {
		key := fmt.Sprintf("%s:%s", v.VersionKey, settings.Common.SecretKey)
		hash := hmac.New(sha256.New, []byte(key)).Sum(nil)
		if v.Signature != hex.EncodeToString(hash) {
			return api.NewGenericError(http.StatusBadRequest, "Invalid signature.")
		}

		return ctr.c.InvalidateVersion(v.VersionKey, settings.Common.CacheTimeout)
	}); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
