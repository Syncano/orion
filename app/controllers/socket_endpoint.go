package controllers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/delicb/gstring"
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/serializers"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/orion/pkg/redisdb"
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

func (ctr *Controller) SocketEndpointList(c echo.Context) error {
	var (
		o []*models.SocketEndpoint
		q *orm.Query
	)

	// Filter by socket if needed.
	mgr := ctr.q.NewSocketEndpointManager(c)
	socketName := c.Param("socket_name")

	if socketName != "" {
		s := &models.Socket{Name: socketName}

		if err := ctr.q.NewSocketManager(c).OneByName(s); err != nil {
			if err == pg.ErrNoRows {
				return api.NewNotFoundError(s)
			}

			return err
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

func matchRegex(c echo.Context, regex *regexp.Regexp, path, detailParam string, list, retrieve echo.HandlerFunc) (handler echo.HandlerFunc, newPath string) {
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

func (ctr *Controller) socketEndpointHandler(h echo.HandlerFunc, requireAuth bool, call map[string]interface{}) echo.HandlerFunc {
	// Check auth when private flag is true.
	private := call["private"]
	if private != nil {
		requireAuth = requireAuth || private.(bool)
	}

	if h == nil {
		// Set default handler based on call type.
		switch call["type"] {
		case socketCallChannel:
			h = ctr.SocketEndpointChannelRun
		case socketCallScript:
			h = ctr.SocketEndpointCodeboxRun
		default:
			panic("unknown call type")
		}
	}

	// Add common channel wrapper.
	if call["type"] == socketCallChannel {
		h = ctr.socketEndpointChannel(h)
	}

	if requireAuth {
		return ctr.InstanceAuth(ctr.RequireAdmin(h))
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
func (ctr *Controller) SocketEndpointMap(c echo.Context) error {
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
		h = ctr.SocketEndpointInvalidate
		p = p[:len(p)-invalidateURISuffixLen]
	} else {
		// /traces/ allows method=GET, always requires auth.
		// /history/ allows method=GET, requires auth if private.
		method = http.MethodGet
		h, p = matchRegex(c, socketEndpointTraceURIRegex, p, "trace_id", ctr.SocketEndpointTraceList, ctr.SocketEndpointTraceRetrieve)
		if h != nil {
			requireAuth = true
		} else {
			h, p = matchRegex(c, socketEndpointHistoryURIRegex, p, "change_id", ctr.SocketEndpointHistoryList, ctr.SocketEndpointHistoryRetrieve)
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

	if err := ctr.q.NewSocketEndpointManager(c).OneByName(o); err != nil {
		if err == pg.ErrNoRows {
			return api.NewNotFoundError(o)
		}

		return err
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
	return ctr.socketEndpointHandler(h, requireAuth, call)(c)
}

func (ctr *Controller) socketEndpointChannel(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Add channel to context for channel socket endpoint type.
		ch := &models.Channel{Name: models.ChannelDefaultName}
		if err := ctr.q.NewChannelManager(c).OneByName(ch); err != nil {
			if err == pg.ErrNoRows {
				return api.NewNotFoundError(ch)
			}

			return err
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

func (ctr *Controller) createSocketTraceDBCtx(c echo.Context, o interface{}) *redisdb.DBCtx {
	return ctr.redis.DB().Model(o, map[string]interface{}{
		"instance":        c.Get(settings.ContextInstanceKey).(*models.Instance),
		"socket_endpoint": c.Get(contextSocketEndpointKey).(*models.SocketEndpoint),
	})
}

func (ctr *Controller) SocketEndpointTraceList(c echo.Context) error {
	var o []*models.SocketTrace

	paginator := &PaginatorRedis{DBCtx: ctr.createSocketTraceDBCtx(c, &o), SkippedFields: []string{"result", "args"}}
	cursor := paginator.CreateCursor(c, false)

	r, err := Paginate(c, cursor, (*models.SocketTrace)(nil), serializers.SocketTraceListSerializer{}, paginator)
	if err != nil {
		return err
	}

	return api.Render(c, http.StatusOK, serializers.CreatePage(c, r, nil))
}

func (ctr *Controller) SocketEndpointTraceRetrieve(c echo.Context) error {
	o := &models.SocketTrace{}

	v, ok := api.IntGet(c, "trace_id")
	if !ok {
		return api.NewNotFoundError(o)
	}

	if err := ctr.createSocketTraceDBCtx(c, o).Find(v); err != nil {
		if err == redisdb.ErrNotFound {
			return api.NewNotFoundError(o)
		}

		return err
	}

	return api.Render(c, http.StatusOK, serializers.SocketTraceSerializer{}.Response(o))
}

func createEndpointCacheKey(instanceID int, endpointName, hash string) string {
	return fmt.Sprintf("%d:cache:s:%s:%s", instanceID, endpointName, hash)
}

func (ctr *Controller) SocketEndpointInvalidate(c echo.Context) error {
	endpoint := c.Get(contextSocketEndpointKey).(*models.SocketEndpoint)
	s := &models.Socket{ID: endpoint.SocketID}

	if err := ctr.q.NewSocketManager(c).OneByID(s); err != nil {
		if err == pg.ErrNoRows {
			return api.NewNotFoundError(s)
		}

		return err
	}

	cacheKey := createEndpointCacheKey(c.Get(settings.ContextInstanceKey).(*models.Instance).ID, endpoint.Name, s.Hash())
	ctr.redis.Client().Del(cacheKey)

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

func (ctr *Controller) SocketEndpointChannelRun(c echo.Context) error {
	room := c.Get(contextChannelRoomKey).(string)
	return ctr.changeSubscribe(c, &room)
}

func (ctr *Controller) SocketEndpointHistoryList(c echo.Context) error {
	return ctr.changeList(c, c.Get(contextChannelRoomKey).(string))
}

func (ctr *Controller) SocketEndpointHistoryRetrieve(c echo.Context) error {
	o := &models.Change{}

	v, ok := api.IntGet(c, "change_id")
	if !ok {
		return api.NewNotFoundError(o)
	}

	o.ID = v

	return ctr.changeRetrieve(c, c.Get(contextChannelRoomKey).(string), o)
}
