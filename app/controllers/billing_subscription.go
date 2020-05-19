package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/settings"
)

const (
	contextSubscriptionKey = "subscription"
	contextAdminLimitKey   = "admin_limit"
)

func (ctr *Controller) InstanceSubscriptionContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		instance := c.Get(settings.ContextInstanceKey).(*models.Instance)
		o := &models.Subscription{AdminID: instance.OwnerID}
		limit := &models.AdminLimit{AdminID: instance.OwnerID}

		if ctr.q.NewSubscriptionManager(c).OneActiveForAdmin(o, time.Now()) != nil ||
			ctr.q.NewAdminLimitManager(c).OneForAdmin(limit) != nil {
			return api.NewGenericError(http.StatusForbidden, "No active subscription.")
		}

		c.Set(contextSubscriptionKey, o)
		c.Set(contextAdminLimitKey, limit)
		c.Set(api.ContextInstanceRateLimitKey, limit.RateLimit(o))

		return next(c)
	}
}
