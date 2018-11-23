package serializers

import (
	"github.com/labstack/echo"

	"github.com/Syncano/orion/app/models"
)

var httpAllMethods = []string{echo.POST, echo.PUT, echo.PATCH, echo.GET, echo.DELETE}

// SocketEndpointResponse ...
type SocketEndpointResponse struct {
	Name           string      `json:"name"`
	AllowedMethods []string    `json:"allowed_methods"`
	Metadata       models.JSON `json:"metadata"`
}

// SocketEndpointSerializer ...
type SocketEndpointSerializer struct{}

// Response ...
func (s SocketEndpointSerializer) Response(i interface{}) interface{} {
	o := i.(*models.SocketEndpoint)
	var allowedMethods []string
	for _, call := range o.Calls.Get().([]interface{}) {
		methods := call.(map[string]interface{})["methods"].([]interface{})

		if len(methods) == 1 && methods[0] == "*" {
			allowedMethods = httpAllMethods
		} else {
			for _, meth := range methods {
				allowedMethods = append(allowedMethods, meth.(string))
			}
		}
	}
	return &SocketEndpointResponse{
		Name:           o.Name,
		AllowedMethods: allowedMethods,
		Metadata:       o.Metadata,
	}
}
