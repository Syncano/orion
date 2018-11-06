package api

import (
	"strconv"

	"github.com/labstack/echo"
)

// IntParam ...
func IntParam(c echo.Context, name string, o Verboser) (int, bool) {
	v, err := strconv.Atoi(c.Param(name))
	if err != nil {
		return 0, false
	}
	return v, true
}
