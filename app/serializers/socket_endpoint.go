package serializers

import (
	"net/http"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/v2/database/fields"
)

var httpAllMethods = []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodGet, http.MethodDelete}

type SocketEndpointResponse struct {
	Name           string      `json:"name"`
	AllowedMethods []string    `json:"allowed_methods"`
	Metadata       fields.JSON `json:"metadata"`
}

type SocketEndpointSerializer struct{}

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
