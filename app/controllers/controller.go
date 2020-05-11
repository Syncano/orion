package controllers

import (
	"reflect"
	"strings"

	"github.com/go-pg/pg/v9/orm"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/query"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/orion/app/tasks"
	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/celery"
	"github.com/Syncano/orion/pkg/log"
	"github.com/Syncano/orion/pkg/storage"
	broker "github.com/Syncano/syncanoapis/gen/go/syncano/codebox/broker/v1"
)

type Controller struct {
	c         *cache.Cache
	db        *storage.Database
	fs        *storage.Storage
	redis     *storage.Redis
	q         *query.Factory
	cel       *celery.Celery
	brokerCli broker.ScriptRunnerClient
	log       *log.Logger
}

func New(db *storage.Database, fs *storage.Storage, redis *storage.Redis, c *cache.Cache, cel *celery.Celery, logger *log.Logger) (*Controller, error) {
	conn, err := grpc.Dial(settings.Socket.CodeboxAddr,
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(settings.MaxGRPCMessageSize)),
		grpc.WithUnaryInterceptor(
			grpc_retry.UnaryClientInterceptor(grpc_retry.WithMax(settings.Socket.CodeboxRetry)),
		),
		grpc.WithStreamInterceptor(
			grpc_retry.StreamClientInterceptor(grpc_retry.WithMax(settings.Socket.CodeboxRetry)),
		),
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
	db.AddModelSoftDeleteHook(storage.AnyModel, ctr.liveObjectSoftDeleteHook)

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

	return ctr, nil
}

func (ctr *Controller) Redis() *storage.Redis {
	return ctr.redis
}

func (ctr *Controller) cacheSaveHook(c storage.DBContext, db orm.DB, created bool, m interface{}) error {
	if created {
		return nil
	}

	ctr.c.ModelCacheInvalidate(db, m)

	return nil
}

func (ctr *Controller) cacheDeleteHook(c storage.DBContext, db orm.DB, m interface{}) error {
	ctr.c.ModelCacheInvalidate(db, m)
	return nil
}

func (ctr *Controller) liveObjectSoftDeleteHook(c storage.DBContext, db orm.DB, m interface{}) error {
	table := orm.GetTable(reflect.TypeOf(m).Elem())
	n := strings.Split(string(table.FullName), ".")
	modelName := strings.ReplaceAll(n[len(n)-1], "_", ".")

	objectPK := table.PKs[0].Value(reflect.ValueOf(m).Elem()).Interface()

	ctr.db.AddDBCommitHook(db, func() error {
		return tasks.NewDeleteLiveObjectTask(
			c.Get(settings.ContextInstanceKey).(*models.Instance).ID,
			modelName, objectPK,
		).Publish(ctr.cel)
	})

	return nil
}
