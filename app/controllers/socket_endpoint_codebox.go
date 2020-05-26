package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
	redis_cache "github.com/go-redis/cache/v7"
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	broker "github.com/Syncano/syncanoapis/gen/go/syncano/codebox/broker/v1"
	"github.com/Syncano/syncanoapis/gen/go/syncano/codebox/lb/v1"
	"github.com/Syncano/syncanoapis/gen/go/syncano/codebox/script/v1"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/serializers"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/pkg-go/util"
)

type empty struct{}

var (
	socketEndpointProtectedKeys         = []string{"api_key", "_api_key", "user_key", "_user_key"}
	socketEndpointDisallowedMetaHeaders = map[string]empty{
		"X-FORWARDED-FOR": {}, "X-FORWARDED-PROTO": {}, "X-FORWARDED-PORT": {}, "X-FORWARDED-HOST": {}, "X-ORIGINAL-FORWARDED-FOR": {}, "HTTP_X_SCHEME": {},
		"X-USER-KEY": {}, "AUTHORIZATION": {}, "X-API-KEY": {}, "HOST-TYPE": {},
		"X-REAL-IP": {},
	}

	codeToStatus = map[int32]string{
		0:   models.TraceStatusSuccess,
		124: models.TraceStatusTimeout,
	}
)

const (
	// Timeout is a default timeout for codebox grpc.
	codeboxTimeout = 8 * time.Minute

	socketEndpointTokenDuration = 10 * time.Minute
	socketEndpointTraceType     = "socket_endpoint"
	codeboxTraceFormat          = `{"id":%d,"instance_id":%d,"obj_id":%d,"obj_name":"%s","type":"%s","socket":"%s"}`
	getSkipCache                = "__skip_cache"
)

type socketEndpointFile struct {
	ContentType string
	Filename    string
	Data        []byte
}

func createCodeboxTraceKey(typ string, inst *models.Instance, sock *models.Socket, endpoint *models.SocketEndpoint, trace *models.SocketTrace) string {
	return fmt.Sprintf(codeboxTraceFormat, trace.ID, inst.ID, endpoint.ID, endpoint.Name, typ, sock.Name)
}

func prepareSocketEndpointPayload(c echo.Context) (payload map[string]interface{}, files map[string]*socketEndpointFile, err error) {
	payload = make(map[string]interface{})
	files = make(map[string]*socketEndpointFile)

	if values, err := c.FormParams(); err == nil {
		// Form.
		for k, vals := range values {
			if len(k) > 2 && strings.HasSuffix(k, "[]") {
				payload[k[:len(k)-2]] = vals
			} else {
				payload[k] = vals[0]
			}
		}
	}

	// FormData.
	if f, err := c.MultipartForm(); err == nil {
		for k, vals := range f.File {
			file := vals[0]

			if f, err := file.Open(); err == nil {
				if buf, e := ioutil.ReadAll(f); e == nil {
					files[k] = &socketEndpointFile{
						Filename:    file.Filename,
						ContentType: file.Header.Get("Content-Type"),
						Data:        buf,
					}
				}
			}
		}
	} else if data, err := api.ParsedData(c); err != echo.ErrUnsupportedMediaType {
		// JSON.
		if err == nil {
			for k, vals := range data {
				payload[k] = vals
			}
		} else if err != io.EOF {
			return nil, nil, api.NewBadRequestError("Parsing payload failure: invalid JSON.")
		}
	}

	for _, k := range socketEndpointProtectedKeys {
		delete(payload, k)
	}

	return payload, files, nil
}

func prepareSocketEndpointMeta(c echo.Context, inst *models.Instance, sock *models.Socket, endpoint *models.SocketEndpoint) map[string]interface{} {
	req := c.Request()
	rm := map[string]interface{}{
		"PATH_INFO":      req.URL.Path,
		"REMOTE_ADDR":    c.RealIP(),
		"REQUEST_METHOD": req.Method,
		"HTTP_HOST":      req.Host,
	}

	// Endpoint metadata.
	endpointMeta := endpoint.Metadata.Get().(map[string]interface{})
	metadata := interface{}(endpointMeta)

	if m, ok := endpointMeta[req.Method]; ok {
		metadata = m
	}

	meta := map[string]interface{}{
		"request":     rm,
		"metadata":    metadata,
		"executed_by": "socket_endpoint",
		"executor":    endpoint.Name,
		"instance":    inst.Name,
		"socket":      sock.Name,
		"token":       createAuthToken(inst, socketEndpointTokenDuration),
		"api_host":    settings.API.Host,
	}

	if len(settings.API.SpaceHost) > 0 {
		meta["space_host"] = settings.API.SpaceHost
	}

	for h, v := range req.Header {
		h = strings.ToUpper(h)
		if _, ok := socketEndpointDisallowedMetaHeaders[h]; ok {
			continue
		}

		h = fmt.Sprintf("HTTP_%s", strings.ReplaceAll(h, "-", "_"))
		rm[h] = v[0]
	}

	// Inject admin.
	if k := c.Get(settings.ContextAdminKey); k != nil {
		admin := k.(*models.Admin)
		meta["admin"] = map[string]interface{}{
			"id":    admin.ID,
			"email": admin.Email,
		}
	}

	// Inject user.
	if k := c.Get(settings.ContextUserKey); k != nil {
		user := k.(*models.User)
		meta["user"] = map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"user_key": user.Key,
		}
	}

	return meta
}

func prepareSocketEndpointConfig(inst *models.Instance, sock *models.Socket) map[string]interface{} {
	cfg := make(map[string]interface{})
	for k, v := range inst.Config.Get().(map[string]interface{}) {
		cfg[k] = v
	}

	for k, v := range sock.Config.Get().(map[string]interface{}) {
		cfg[k] = v
	}

	return cfg
}

func (ctr *Controller) sendCodeboxRequest(ctx context.Context, c echo.Context, inst *models.Instance, sock *models.Socket,
	endpoint *models.SocketEndpoint) (broker.ScriptRunner_RunClient, *models.SocketTrace, error) {
	call := c.Get(contextSocketEndpointCallKey).(map[string]interface{})
	sub := c.Get(contextSubscriptionKey).(*models.Subscription)

	var environmentHash, environmentURL string

	if sock.EnvironmentID != 0 {
		environment := &models.SocketEnvironment{ID: sock.EnvironmentID}
		if err := ctr.q.NewSocketEnvironmentManager(c).OneByID(environment); err != nil {
			if err == pg.ErrNoRows {
				return nil, nil, api.NewNotFoundError(endpoint)
			}

			return nil, nil, err
		}

		environmentHash = environment.Hash()
		environmentURL = environment.URL(ctr.fs.Default())
	}

	// Process payload.
	payload, files, err := prepareSocketEndpointPayload(c)
	if err != nil {
		return nil, nil, err
	}

	payloadBytes, err := jsoniter.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	// Process meta.
	meta := prepareSocketEndpointMeta(c, inst, sock, endpoint)

	metaBytes, err := jsoniter.Marshal(meta)
	if err != nil {
		return nil, nil, err
	}

	// Process meta.
	config := prepareSocketEndpointConfig(inst, sock)

	configBytes, err := jsoniter.Marshal(config)
	if err != nil {
		return nil, nil, err
	}

	// Prepare trace.
	trace := &models.SocketTrace{
		Meta: meta["request"].(map[string]interface{}),
		Args: payload,
	}

	if e := ctr.createSocketTraceDBCtx(c, trace).Save(nil); e != nil {
		return nil, nil, e
	}

	async := settings.Socket.DefaultAsync
	timeout := int64(settings.Socket.DefaultTimeout)
	mcpu := settings.Socket.DefaultMCPU

	if v, ok := call["async"]; ok {
		async = uint32(v.(float64))
		meta["async"] = async
	}

	if v, ok := call["timeout"]; ok {
		timeout = int64(v.(float64) * 1000)
	}

	if v, ok := call["mcpu"]; ok {
		mcpu = uint32(v.(float64))
	}

	// Prepare request.
	instanceID := strconv.Itoa(inst.ID)

	ctx, _ = api.AddRequestID(ctx, c)

	stream, err := ctr.brokerCli.Run(ctx, grpc.WaitForReady(true))
	if err != nil {
		return nil, nil, err
	}

	// Broker Meta
	req := &broker.RunRequest{
		Value: &broker.RunRequest_Meta{
			Meta: &broker.RunMeta{
				Files:          sock.Files(ctr.fs.Default()),
				EnvironmentUrl: environmentURL,
				Trace:          []byte(createCodeboxTraceKey(socketEndpointTraceType, inst, sock, endpoint, trace)),
				TraceId:        uint64(trace.ID),
				Sync:           true,
			},
		},
	}
	if err := stream.Send(req); err != nil {
		return nil, nil, err
	}

	// LB Meta
	req = &broker.RunRequest{
		Value: &broker.RunRequest_LbMeta{
			LbMeta: &lb.RunMeta{
				ConcurrencyKey:   instanceID,
				ConcurrencyLimit: int32(c.Get(contextAdminLimitKey).(*models.AdminLimit).CodeboxConcurrency(sub)),
			},
		},
	}
	if err := stream.Send(req); err != nil {
		return nil, nil, err
	}

	// Script Meta
	req = &broker.RunRequest{
		Value: &broker.RunRequest_ScriptMeta{
			ScriptMeta: &script.RunMeta{
				Runtime:     call["runtime"].(string),
				SourceHash:  sock.Hash(),
				UserId:      instanceID,
				Environment: environmentHash,

				Options: &script.RunMeta_Options{
					Entrypoint:  endpoint.Entrypoint(call),
					OutputLimit: uint32(settings.Socket.MaxResultSize),
					Timeout:     timeout,
					Async:       async,
					Mcpu:        mcpu,
					Config:      configBytes,
					Meta:        metaBytes,
				},
			},
		},
	}
	if err := stream.Send(req); err != nil {
		return nil, nil, err
	}

	// Script Chunks
	req = &broker.RunRequest{
		Value: &broker.RunRequest_ScriptChunk{
			ScriptChunk: &script.RunChunk{
				Data: payloadBytes,
				Type: script.RunChunk_ARGS,
			},
		},
	}
	if err := stream.Send(req); err != nil {
		return nil, nil, err
	}

	// Add files to request.
	for n, f := range files {
		req = &broker.RunRequest{
			Value: &broker.RunRequest_ScriptChunk{
				ScriptChunk: &script.RunChunk{
					Name:        n,
					Filename:    f.Filename,
					ContentType: f.ContentType,
					Data:        f.Data,
				},
			},
		}
		if err := stream.Send(req); err != nil {
			return nil, nil, err
		}
	}

	if err := stream.CloseSend(); err != nil {
		return nil, nil, err
	}

	return stream, trace, err
}

func processCodeboxResponse(stream broker.ScriptRunner_RunClient, trace *models.SocketTrace) error {
	result, err := stream.Recv()
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.ResourceExhausted {
				trace.Status = models.TraceStatusBlocked
				trace.ExecutedAt = time.Now().UTC()

				return nil
			}
		}

		return err
	}
	// Read until all chunks arrive.
	for {
		chunk, e := stream.Recv()
		if e != nil {
			if e != io.EOF {
				return e
			}

			break
		}

		result.Response.Content = append(result.Response.Content, chunk.Response.Content...)
	}

	ret := map[string]interface{}{
		"stdout": json.RawMessage(util.ToQuoteJSON(result.Stdout)),
		"stderr": json.RawMessage(util.ToQuoteJSON(result.Stderr)),
	}

	res := result.GetResponse()
	if res != nil {
		ret["response"] = map[string]interface{}{
			"status":       int(res.StatusCode),
			"content_type": res.ContentType,
			"content":      res.Content,
			"headers":      res.GetHeaders(),
		}
	}

	trace.ExecutedAt = time.Unix(0, result.Time)
	trace.Result = ret
	trace.Duration = int(result.Took)
	trace.Status = models.TraceStatusFailure

	if s, ok := codeToStatus[result.Code]; ok {
		trace.Status = s
	}

	return nil
}

func (ctr *Controller) SocketEndpointCodeboxRun(c echo.Context) error {
	instance := c.Get(settings.ContextInstanceKey).(*models.Instance)
	endpoint := c.Get(contextSocketEndpointKey).(*models.SocketEndpoint)
	call := c.Get(contextSocketEndpointCallKey).(map[string]interface{})
	socket := &models.Socket{ID: endpoint.SocketID}

	if err := ctr.q.NewSocketManager(c).OneByID(socket); err != nil {
		if err == pg.ErrNoRows {
			return api.NewNotFoundError(socket)
		}

		return err
	}

	var (
		trace  *models.SocketTrace
		stream broker.ScriptRunner_RunClient
		err    error
	)

	// Process caching.
	if v, ok := call["cache"]; ok && c.QueryParam(getSkipCache) != "1" {
		cacheTimeout := v.(float64)
		cacheKey := createEndpointCacheKey(c.Get(settings.ContextInstanceKey).(*models.Instance).ID, endpoint.Name, socket.Hash())

		if ctr.c.Codec().Get(cacheKey, trace) == nil {
			return serializers.SocketTraceSerializer{}.Render(c, trace)
		}

		defer func() {
			if trace != nil {
				ctr.c.Codec().Set(&redis_cache.Item{ // nolint: errcheck
					Key:        cacheKey,
					Object:     trace,
					Expiration: time.Duration(cacheTimeout),
				})
			}
		}()
	}

	// Process request.
	ctx, cancel := context.WithTimeout(c.Request().Context(), codeboxTimeout)
	defer cancel()

	stream, trace, err = ctr.sendCodeboxRequest(ctx, c, instance, socket, endpoint)
	if err != nil {
		return fmt.Errorf("error sending codebox request: %w", err)
	}

	if err := processCodeboxResponse(stream, trace); err != nil {
		return fmt.Errorf("error processing codebox response: %w", err)
	}

	return serializers.SocketTraceSerializer{}.Render(c, trace)
}
