package serializers

import (
	"github.com/labstack/echo"

	"github.com/Syncano/orion/app/api"
)

// CreatePage ...
func CreatePage(c echo.Context, objects []api.RawMessage, properties map[string]interface{}) map[string]interface{} {
	ret := make(map[string]interface{})

	for k, v := range properties {
		ret[k] = v
	}

	ret["next"] = c.Get("next")
	ret["prev"] = c.Get("prev")

	// If objects are empty, serialize as empty array.
	if len(objects) == 0 {
		objects = make([]api.RawMessage, 0)
	}
	ret["objects"] = objects
	return ret
}
