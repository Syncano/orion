package controllers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/delicb/gstring"
	"github.com/go-pg/pg/orm"
	"github.com/labstack/echo"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/query"
	"github.com/Syncano/orion/app/serializers"
	"github.com/Syncano/orion/pkg/redisdb"
	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/storage"
	"github.com/Syncano/orion/pkg/util"
)

const (
	contextSocketEndpointKey     = "socket_endpoint"
	contextSocketEndpointCallKey = "socket_endpoint_call"
	contextChannelRoomKey        = "channel_room"
	invalidateURISuffix          = "/invalidate"
	invalidateURISuffixLen       = len(invalidateURISuffix)
	socketCallChannel            = "channel"
	socketCallScript             = "script"
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

func socketEndpointHandler(h echo.HandlerFunc, requireAuth bool, call map[string]interface{}) echo.HandlerFunc {
	// Check auth when private flag is true.
	private := call["private"]
	if private != nil {
		requireAuth = requireAuth || private.(bool)
	}

	if h == nil {
		// Set default handler based on call type.
		switch call["type"] {
		case socketCallChannel:
			h = SocketEndpointChannelRun
		case socketCallScript:
			h = SocketEndpointCodeboxRun
		default:
			panic("unknown call type")
		}
	}

	// Add common channel wrapper.
	if call["type"] == socketCallChannel {
		h = socketEndpointChannel(h)
	}

	if requireAuth {
		return InstanceAuth(RequireAdmin(h))
	}
	return h
}

// SocketEndpointMap maps socket endpoints to handlers.
// for call == "script":
//   /x/            -> require auth if private, SocketEndpointCodeboxRun
//   /x/invalidate/ -> require auth if private, SocketEndpointInvalidate
//   /x/traces/     -> require auth, SocketEndpointTraceList / SocketEndpointTraceRetrieve
//
// for call == "channel":
//   /x/            -> require auth if private, SocketEndpointChannelRun
//   /x/history/    -> require auth if private, SocketEndpointHistoryList / SocketEndpointHistoryRetrieve
func SocketEndpointMap(c echo.Context) error {
	p := c.Param("*")
	if !strings.HasSuffix(p, "/") {
		return echo.ErrNotFound
	}
	p = p[:len(p)-1]

	var (
		h           echo.HandlerFunc
		method      string
		requireAuth bool
	)
	callType := socketCallScript

	// Match invalidate, history and traces endpoints first.
	if strings.HasSuffix(p, invalidateURISuffix) {
		// /invalidate/ allows method=POST, requires auth if private.
		method = http.MethodPost
		requireAuth = false
		h = SocketEndpointInvalidate
		p = p[:len(p)-invalidateURISuffixLen]
	} else {
		// /traces/ allows method=GET, always requires auth.
		// /history/ allows method=GET, requires auth if private.
		method = http.MethodGet
		h, p = matchRegex(c, socketEndpointTraceURIRegex, p, "trace_id", SocketEndpointTraceList, SocketEndpointTraceRetrieve)
		if h != nil {
			requireAuth = true
		} else {
			h, p = matchRegex(c, socketEndpointHistoryURIRegex, p, "change_id", SocketEndpointHistoryList, SocketEndpointHistoryRetrieve)
			callType = socketCallChannel
			requireAuth = false
		}
	}

	// Check request method.
	if h != nil && c.Request().Method != method {
		return echo.ErrMethodNotAllowed
	}

	// Add socket endpoint to context.
	name := fmt.Sprintf("%s/%s", c.Param("socket_name"), p)
	o := &models.SocketEndpoint{
		Name: name,
	}
	if query.NewSocketEndpointManager(c).OneByName(o) != nil {
		return api.NewNotFoundError(o)
	}
	c.Set(contextSocketEndpointKey, o)

	// Add call to context and check expected call type.
	call := o.MatchCall(c.Request().Method)
	c.Set(contextSocketEndpointCallKey, call)

	if call == nil {
		return echo.ErrMethodNotAllowed
	}
	if h != nil && callType != call["type"] {
		return echo.ErrNotFound
	}

	// Process Socket Endpoint handler.
	return socketEndpointHandler(h, requireAuth, call)(c)
}

func socketEndpointChannel(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Add channel to context for channel socket endpoint type.
		ch := &models.Channel{Name: models.ChannelDefaultName}
		if query.NewChannelManager(c).OneByName(ch) != nil {
			return api.NewNotFoundError(ch)
		}
		c.Set(contextChannelKey, ch)

		room, err := createChannelRoom(c)
		if err != nil {
			return err
		}
		c.Set(contextChannelRoomKey, room)

		return next(c)
	}
}

func createSocketTraceDBCtx(c echo.Context, o interface{}) *redisdb.DBCtx {
	return storage.RedisDB().Model(o, map[string]interface{}{
		"instance":        c.Get(settings.ContextInstanceKey).(*models.Instance),
		"socket_endpoint": c.Get(contextSocketEndpointKey).(*models.SocketEndpoint),
	})
}

// SocketEndpointTraceList ...
func SocketEndpointTraceList(c echo.Context) error {
	var o []*models.SocketTrace

	paginator := &PaginatorRedis{DBCtx: createSocketTraceDBCtx(c, &o), SkippedFields: []string{"result", "args"}}
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

	if createSocketTraceDBCtx(c, o).Find(v) != nil {
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

func createChannelRoom(c echo.Context) (string, error) {
	format := c.Get(contextSocketEndpointCallKey).(map[string]interface{})["channel"].(string)
	// Prepare context for room formatting.
	ctx := make(map[string]interface{})
	for q, v := range c.QueryParams() {
		ctx[q] = v[0]
	}
	delete(ctx, "user")
	if k := c.Get(settings.ContextUserKey); k != nil {
		ctx["user"] = k.(*models.User).Username
	}

	room := gstring.Sprintm(format, ctx)
	if strings.Contains(room, "%MISSING") {
		if _, ok := ctx["user"]; !ok && strings.Contains(format, "{user}") {
			return "", api.NewPermissionDeniedError()
		}
		return "", api.NewGenericError(http.StatusForbidden, "Channel format not satisfied.")
	}
	return room, nil
}

// SocketEndpointChannelRun ...
func SocketEndpointChannelRun(c echo.Context) error {
	room := c.Get(contextChannelRoomKey).(string)
	return changeSubscribe(c, &room)
}

// SocketEndpointHistoryList ...
func SocketEndpointHistoryList(c echo.Context) error {
	return changeList(c, c.Get(contextChannelRoomKey).(string))
}

// SocketEndpointHistoryRetrieve ...
func SocketEndpointHistoryRetrieve(c echo.Context) error {
	o := &models.Change{}
	v, ok := api.IntGet(c, "change_id")
	if !ok {
		return api.NewNotFoundError(o)
	}
	o.ID = v

	return changeRetrieve(c, c.Get(contextChannelRoomKey).(string), o)
}
