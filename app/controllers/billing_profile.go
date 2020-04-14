package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/query"
)

func BillingCheck(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sub := c.Get(contextSubscriptionKey).(*models.Subscription)
		mgr := query.NewProfileManager(c)

		status, err := mgr.GetBillingStatus(sub)
		if err != nil {
			return err
		}

		var str string

		switch status {
		case query.BillingStatusNoActiveSubscription:
			str = "No active subscription."
		case query.BillingStatusHardLimitExceeded:
			str = "Hard limit reached."
		case query.BillingStatusFreeLimitExceeded:
			str = "Free limits exceeded."
		case query.BillingStatusOverdueInvoices:
			str = "Account blocked due to overdue invoices."
		default:
			return next(c)
		}

		return api.NewGenericError(http.StatusForbidden, str)
	}
}
