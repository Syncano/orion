package controllers

import (
	"reflect"
	"strings"

	"github.com/go-pg/pg/v9/orm"
	"github.com/labstack/echo/v4"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/query"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/orion/app/tasks"
	"github.com/Syncano/pkg-go/v2/celery"
	"github.com/Syncano/pkg-go/v2/database"
	"github.com/Syncano/pkg-go/v2/database/fields"
	"github.com/Syncano/pkg-go/v2/log"
	"github.com/Syncano/pkg-go/v2/rediscache"
	"github.com/Syncano/pkg-go/v2/rediscli"
	"github.com/Syncano/pkg-go/v2/storage"
	broker "github.com/Syncano/syncanoapis/gen/go/syncano/codebox/broker/v1"
)

type Controller struct {
	c         *rediscache.Cache
	db        *database.DB
	fs        *storage.Storage
	redis     *rediscli.Redis
	q         *query.Factory
	cel       *celery.Celery
	brokerCli broker.ScriptRunnerClient
	log       *log.Logger
}

func New(db *database.DB, fs *storage.Storage, redis *rediscli.Redis, c *rediscache.Cache, cel *celery.Celery, logger *log.Logger) (*Controller, error) {
	conn, err := grpc.Dial(settings.Socket.CodeboxAddr,
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(settings.MaxGRPCMessageSize)),
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{}),
	)
	if err != nil {
		return nil, err
	}

	ctr := &Controller{
		c:         c,
		db:        db,
		fs:        fs,
		redis:     redis,
		q:         query.NewFactory(db, c),
		cel:       cel,
		brokerCli: broker.NewScriptRunnerClient(conn),
		log:       logger,
	}

	// DataObject cleanup.
	db.AddModelDeleteHook((*models.DataObject)(nil), ctr.dataObjectDeleteHook)

	// Triggers.
	db.AddModelSoftDeleteHook((*models.DataObject)(nil), ctr.dataObjectSoftDeleteTriggerHook)

	// LiveObject cleanup.
	// TODO: InstanceIndicator post save hook after live obj delete is done.
	db.AddModelSoftDeleteHook(database.AnyModel, ctr.liveObjectSoftDeleteHook)

	// Cache invalidate hooks.
	for _, model := range []interface{}{
		(*models.AdminInstanceRole)(nil),
		(*models.Admin)(nil),
		(*models.APIKey)(nil),
		(*models.Profile)(nil),
		(*models.Subscription)(nil),
		(*models.Instance)(nil),
		(*models.Channel)(nil),
		(*models.Class)(nil),
		(*models.Codebox)(nil),
		(*models.SocketEndpoint)(nil),
		(*models.SocketEnvironment)(nil),
		(*models.Socket)(nil),
		(*models.User)(nil),
	} {
		db.AddModelDeleteHook(model, ctr.cacheDeleteHook)
		db.AddModelSoftDeleteHook(model, ctr.cacheDeleteHook)
		db.AddModelSaveHook(model, ctr.cacheSaveHook)
	}

	fields.DateTimeFormat = settings.Common.DateTimeFormat

	return ctr, nil
}

func (ctr *Controller) Redis() *rediscli.Redis {
	return ctr.redis
}

func (ctr *Controller) cacheSaveHook(c database.DBContext, db orm.DB, created bool, m interface{}) error {
	if created {
		return nil
	}

	ctr.c.ModelCacheInvalidate(db, m)

	return nil
}

func (ctr *Controller) cacheDeleteHook(c database.DBContext, db orm.DB, m interface{}) error {
	ctr.c.ModelCacheInvalidate(db, m)
	return nil
}

func (ctr *Controller) liveObjectSoftDeleteHook(c database.DBContext, db orm.DB, m interface{}) error {
	table := orm.GetTable(reflect.TypeOf(m).Elem())
	n := strings.Split(string(table.FullName), ".")
	modelName := strings.ReplaceAll(n[len(n)-1], "_", ".")

	objectPK := table.PKs[0].Value(reflect.ValueOf(m).Elem()).Interface()

	ctr.db.AddDBCommitHook(db, func() error {
		return tasks.NewDeleteLiveObjectTask(
			c.(echo.Context).Get(settings.ContextInstanceKey).(*models.Instance).ID,
			modelName, objectPK,
		).Publish(ctr.cel)
	})

	return nil
}
