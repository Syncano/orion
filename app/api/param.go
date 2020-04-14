package api

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

func IntParam(c echo.Context, name string) (int, bool) {
	v, err := strconv.Atoi(c.Param(name))
	if err != nil {
		return 0, false
	}

	return v, true
}

func IntGet(c echo.Context, name string) (int, bool) {
	v, err := strconv.Atoi(c.Get(name).(string))
	if err != nil {
		return 0, false
	}

	return v, true
}
