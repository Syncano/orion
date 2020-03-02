package controllers

import (
	"crypto/md5" // nolint: gosec
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/serializers"
	"github.com/Syncano/orion/pkg/redisdb"
	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/storage"
)

const (
	contextChannelKey = "channel"
)

var (
	changeIDRegex = regexp.MustCompile(`"id":\s*(\d+)`)
	upgrader      = websocket.Upgrader{
		CheckOrigin: func(*http.Request) bool { return true },
		Error:       func(w http.ResponseWriter, r *http.Request, status int, reason error) {},
	}
)

func createChangeDBCtx(c echo.Context, room string, o interface{}) *redisdb.DBCtx {
	return storage.RedisDB().Model(o, map[string]interface{}{
		"instance": c.Get(settings.ContextInstanceKey).(*models.Instance),
		"channel":  c.Get(contextChannelKey).(*models.Channel),
		"room":     room,
	})
}

func changeList(c echo.Context, room string) error {
	var o []*models.Change

	paginator := &PaginatorRedis{DBCtx: createChangeDBCtx(c, room, &o)}
	cursor := paginator.CreateCursor(c, false)

	r, err := Paginate(c, cursor, (*models.Change)(nil), serializers.ChangeSerializer{}, paginator)
	if err != nil {
		return err
	}

	return api.Render(c, http.StatusOK, serializers.CreatePage(c, r, nil))
}

func changeRetrieve(c echo.Context, room string, o *models.Change) error {
	if createChangeDBCtx(c, room, o).Find(o.ID) != nil {
		return api.NewNotFoundError(o)
	}

	return api.Render(c, http.StatusOK, serializers.ChangeSerializer{}.Response(o))
}

func channelWithRoom(s string, room *string) string {
	if room != nil {
		s += fmt.Sprintf(":%x", md5.Sum([]byte(*room))) // nolint: gosec
	}

	return s
}
func channelPublishLockKey(inst *models.Instance, ch *models.Channel, room *string) string { // nolint - ignore that it is unused for now
	return channelWithRoom(fmt.Sprintf("lock:channel:publish:%d:%d", inst.ID, ch.ID), room)
}
func channelStreamKey(inst *models.Instance, ch *models.Channel, room *string) string {
	return channelWithRoom(fmt.Sprintf("stream:channel:%d:%d", inst.ID, ch.ID), room)
}

func changeSubscribe(c echo.Context, room *string) error {
	isWebSocket := c.QueryParam("transport") == "websocket"
	limit := 1

	if isWebSocket {
		limit = settings.API.ChannelWebSocketLimit
	}

	ch, err := changeSubscribeStream(c, room, limit, settings.API.ChannelSubscribeTimeout)
	if err != nil {
		return err
	}

	if isWebSocket {
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		defer ws.Close()

		for o := range ch {
			if err := ws.WriteMessage(websocket.TextMessage, o); err != nil {
				return err
			}
		}

		return nil
	}

	o := <-ch
	if o != nil {
		return c.JSONBlob(http.StatusOK, o)
	}

	return c.NoContent(http.StatusNoContent)
}

func changeSubscribeStream(c echo.Context, room *string, limit int, timeout time.Duration) (<-chan []byte, error) {
	instance := c.Get(settings.ContextInstanceKey).(*models.Instance)
	channel := c.Get(contextChannelKey).(*models.Channel)
	start := time.Now()
	outCh := make(chan []byte, limit)

	var (
		lastID int
		err    error
	)

	if lastIDStr := c.QueryParam("last_id"); lastIDStr != "" {
		lastID, _ = strconv.Atoi(lastIDStr)

		var o []*models.Change
		if err := createChangeDBCtx(c, *room, &o).List(lastID+1, 0, limit, true, nil); err != nil {
			return nil, err
		}

		var b []byte
		for _, obj := range o {
			b, err = api.Marshal(c, serializers.ChangeSerializer{}.Response(obj))
			outCh <- b

			lastID = obj.ID

			if err != nil {
				return outCh, err
			}
		}
	}

	limit -= len(outCh)
	if limit == 0 {
		return outCh, err
	}

	streamKey := channelStreamKey(instance, channel, room)
	ch := make(chan string, limit)

	if err := storage.RedisPubSub().Subscribe(streamKey, ch); err != nil {
		return outCh, err
	}

	go func() {
		var timer *time.Timer

		for {
			timer = time.NewTimer(timeout - time.Since(start))

			select {
			case o := <-ch:
				// Extract ID from change and check it's value. Using regex to avoid unnecessary JSON unmarshaling.
				if lastID > 0 {
					m := changeIDRegex.FindStringSubmatch(o)
					id, _ := strconv.Atoi(m[1])

					if id <= lastID {
						break
					}

					lastID = 0
				}

				outCh <- []byte(o)
				limit--

				if limit == 0 {
					close(outCh)
					return
				}

			case <-timer.C:
				close(outCh)
				return
			}
		}
	}()

	return outCh, nil
}
