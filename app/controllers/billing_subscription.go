package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/query"
	"github.com/Syncano/orion/pkg/settings"
)

const (
	contextSubscriptionKey = "subscription"
	contextAdminLimitKey   = "admin_limit"
)

func InstanceSubscriptionContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		instance := c.Get(settings.ContextInstanceKey).(*models.Instance)
		o := &models.Subscription{AdminID: instance.OwnerID}
		limit := &models.AdminLimit{AdminID: instance.OwnerID}

		if query.NewSubscriptionManager(c).OneActiveForAdmin(o, time.Now()) != nil ||
			query.NewAdminLimitManager(c).OneForAdmin(limit) != nil {
			return api.NewGenericError(http.StatusForbidden, "No active subscription.")
		}

		c.Set(contextSubscriptionKey, o)
		c.Set(contextAdminLimitKey, limit)
		c.Set(api.ContextRateLimitKey, limit.RateLimit(o))

		return next(c)
	}
}
