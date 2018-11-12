package controllers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-pg/pg/orm"
	"github.com/labstack/echo"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/query"
	"github.com/Syncano/orion/app/serializers"
	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/storage"
	"github.com/Syncano/orion/pkg/util"
)

const (
	contextSocketEndpointKey     = "socket_endpoint"
	contextSocketEndpointCallKey = "socket_endpoint_call"
	invalidateURISuffix          = "/invalidate"
	invalidateURISuffixLen       = len(invalidateURISuffix)
)

var (
	socketEndpointTraceURIRegex   = regexp.MustCompile("/traces(/(?P<trace_id>[^/]+))?$")
	socketEndpointHistoryURIRegex = regexp.MustCompile("/history(/(?P<change_id>[^/]+))?$")
)

// SocketEndpointList ...
func SocketEndpointList(c echo.Context) error {
	var o []*models.SocketEndpoint
	mgr := query.NewSocketEndpointManager(c)

	// Filter by socket if needed.
	socketName := c.Param("socket_name")
	var q *orm.Query
	if socketName != "" {
		s := &models.Socket{Name: socketName}
		if query.NewSocketManager(c).OneByName(s) != nil {
			return api.NewNotFoundError(s)
		}
		q = mgr.ForSocketQ(s, &o)
	} else {
		q = mgr.WithAccessQ(&o)
	}
	paginator := &PaginatorDB{Query: q}
	cursor := paginator.CreateCursor(c, true)

	r, err := Paginate(c, cursor, (*models.SocketEndpoint)(nil), serializers.SocketEndpointSerializer{}, paginator)
	if err != nil {
		return err
	}
	return api.Render(c, http.StatusOK, serializers.CreatePage(c, r, nil))

}

func matchRegex(c echo.Context, regex *regexp.Regexp, path, detailParam string, list echo.HandlerFunc, retrieve echo.HandlerFunc) (echo.HandlerFunc, string) {
	if r := regex.FindStringSubmatch(path); r != nil {
		path = path[:len(path)-len(r[0])]

		params := util.RegexNamedGroups(regex, r)
		for k, v := range params {
			c.Set(k, v)
		}
		if params[detailParam] != "" {
			return retrieve, path
		}
		return list, path
	}
	return nil, path
}

// SocketEndpointMap ...
func SocketEndpointMap(c echo.Context) error {
	p := c.Param("*")
	if !strings.HasSuffix(p, "/") {
		return echo.ErrNotFound
	}
	p = p[:len(p)-1]

	var (
		h      echo.HandlerFunc
		method string
	)
	requireAuth := true
	// Match invalidate, history and traces endpoints first.
	if strings.HasSuffix(p, invalidateURISuffix) {
		h = SocketEndpointInvalidate
		p = p[:len(p)-invalidateURISuffixLen]
		method = "POST"
	} else {
		h, p = matchRegex(c, socketEndpointTraceURIRegex, p, "trace_id", SocketEndpointTraceList, SocketEndpointTraceRetrieve)
		if h == nil {
			h, p = matchRegex(c, socketEndpointHistoryURIRegex, p, "change_id", SocketEndpointHistoryList, SocketEndpointHistoryRetrieve)
		}
		method = "GET"
	}
	if h != nil && c.Request().Method != method {
		return echo.ErrMethodNotAllowed
	}

	name := fmt.Sprintf("%s/%s", c.Param("socket_name"), p)
	o := &models.SocketEndpoint{
		Name: name,
	}
	if query.NewSocketEndpointManager(c).OneByName(o) != nil {
		return api.NewNotFoundError(o)
	}
	c.Set(contextSocketEndpointKey, o)

	// Process Socket Endpoint runner.
	if h == nil {
		requireAuth = false
		call := o.MatchCall(c.Request().Method)

		// Check auth when private flag is true.
		private := call["private"]
		if private != nil && private.(bool) {
			if c.Get(settings.ContextAdminKey) == nil {
				return api.NewPermissionDeniedError()
			}
		}

		// Set handler based on call type.
		if call["type"] == "channel" {
			h = SocketEndpointChannelRun
		} else {
			h = SocketEndpointCodeboxRun
		}
		c.Set(contextSocketEndpointCallKey, call)
	}

	if requireAuth {
		return InstanceAuth(RequireAdmin(h))(c)
	}
	return h(c)
}

// SocketEndpointTraceList ...
func SocketEndpointTraceList(c echo.Context) error {
	var o []*models.SocketTrace
	q := storage.RedisDB().Model(&o, map[string]interface{}{
		"instance_id":        c.Get(settings.ContextInstanceKey).(*models.Instance).ID,
		"socket_endpoint_id": c.Get(contextSocketEndpointKey).(*models.SocketEndpoint).ID,
	})

	paginator := &PaginatorRedis{DBCtx: q, SkippedFields: []string{"result", "args"}}
	cursor := paginator.CreateCursor(c, false)

	r, err := Paginate(c, cursor, (*models.SocketTrace)(nil), serializers.SocketTraceListSerializer{}, paginator)
	if err != nil {
		return err
	}
	return api.Render(c, http.StatusOK, serializers.CreatePage(c, r, nil))
}

// SocketEndpointTraceRetrieve ...
func SocketEndpointTraceRetrieve(c echo.Context) error {
	o := &models.SocketTrace{}
	v, ok := api.IntGet(c, "trace_id")
	if !ok {
		return api.NewNotFoundError(o)
	}

	if storage.RedisDB().Model(o, map[string]interface{}{
		"instance_id":        c.Get(settings.ContextInstanceKey).(*models.Instance).ID,
		"socket_endpoint_id": c.Get(contextSocketEndpointKey).(*models.SocketEndpoint).ID,
	}).Find(v) != nil {
		return api.NewNotFoundError(o)
	}

	return api.Render(c, http.StatusOK, serializers.SocketTraceSerializer{}.Response(o))
}

func createEndpointCacheKey(instanceID int, endpointName, hash string) string {
	return fmt.Sprintf("%d:cache:s:%s:%s", instanceID, endpointName, hash)
}

// SocketEndpointInvalidate ...
func SocketEndpointInvalidate(c echo.Context) error {
	endpoint := c.Get(contextSocketEndpointKey).(*models.SocketEndpoint)
	s := &models.Socket{ID: endpoint.SocketID}
	if query.NewSocketManager(c).OneByID(s) != nil {
		return api.NewNotFoundError(s)
	}

	cacheKey := createEndpointCacheKey(c.Get(settings.ContextInstanceKey).(*models.Instance).ID, endpoint.Name, s.Hash())
	storage.Redis().Del(cacheKey)
	return c.NoContent(http.StatusNoContent)
}

// SocketEndpointChannelRun ...
func SocketEndpointChannelRun(c echo.Context) error {
	return api.Render(c, http.StatusOK, map[string]string{
		"channel": "cba",
	})
}

// SocketEndpointHistoryList ...
func SocketEndpointHistoryList(c echo.Context) error {
	return api.Render(c, http.StatusOK, map[string]string{
		"history": "cba",
	})
}

// SocketEndpointHistoryRetrieve ...
func SocketEndpointHistoryRetrieve(c echo.Context) error {
	return api.Render(c, http.StatusOK, map[string]string{
		"history": "retrieve",
	})
}
