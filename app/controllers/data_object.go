package controllers

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/serializers"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/pkg-go/database"
	"github.com/Syncano/pkg-go/database/manager"
	"github.com/Syncano/pkg-go/storage"
	"github.com/Syncano/pkg-go/util"
)

func (ctr *Controller) DataObjectCreate(c echo.Context) error {
	// TODO: #16 Object updates
	// o.Data.Set(map[string]string{ // nolint: errcheck
	// 	"abc": "aa",
	// })
	return api.NewPermissionDeniedError()
}

func (ctr *Controller) DataObjectList(c echo.Context) error {
	var o []*models.DataObject

	mgr := ctr.q.NewDataObjectManager(c)
	props := make(map[string]interface{})
	class := c.Get(contextClassKey).(*models.Class)

	// Prepare query.
	q := mgr.ForClassQ(class, &o)

	if _, e := c.QueryParams()["query"]; e {
		var err error

		q, err = NewDataObjectQuery(class.FilterFields()).Parse(ctr.q, c, q)
		if err != nil {
			return err
		}
	}

	// Check if include_count is defined, if so add count estimate.
	if util.IsTrue(c.QueryParam("include_count")) {
		if count, err := mgr.CountEstimate(q); err == nil {
			props["objects_count"] = count
		}
	}

	// Prepare pagination.
	var paginator Paginator

	if isValidOrderedPagination(c.QueryParam(orderByQuery)) {
		paginator = &PaginatorOrderedDB{PaginatorDB: &PaginatorDB{Query: q}, OrderFields: class.OrderFields()}
	} else {
		paginator = &PaginatorDB{Query: q}
	}

	cursor := paginator.CreateCursor(c, true)

	// Return paginated results.
	serializer := serializers.DataObjectSerializer{Class: class}

	r, err := Paginate(c, cursor, (*models.DataObject)(nil), serializer, paginator)
	if err != nil {
		return err
	}

	return api.Render(c, http.StatusOK, serializers.CreatePage(c, r, props))
}

func detailDataObject(c echo.Context) *models.DataObject {
	o := &models.DataObject{}

	v, ok := api.IntParam(c, "object_id")
	if !ok {
		return nil
	}

	o.ID = v

	return o
}

func (ctr *Controller) DataObjectRetrieve(c echo.Context) error {
	o := detailDataObject(c)
	if o == nil {
		return api.NewNotFoundError(o)
	}

	class := c.Get(contextClassKey).(*models.Class)

	if err := ctr.q.NewDataObjectManager(c).ForClassByIDQ(class, o).Select(); err != nil {
		if err == pg.ErrNoRows {
			return api.NewNotFoundError(o)
		}

		return err
	}

	serializer := serializers.DataObjectSerializer{Class: class}

	return api.Render(c, http.StatusOK, serializer.Response(o))
}

func (ctr *Controller) DataObjectUpdate(c echo.Context) error {
	// TODO: #16 Object updates
	mgr := ctr.q.NewDataObjectManager(c)

	o := detailDataObject(c)
	if o == nil {
		return api.NewNotFoundError(o)
	}

	// Prepare virtual map.
	class := c.Get(contextClassKey).(*models.Class)
	schema := class.ComputedSchema()
	virt := make(map[string]models.StateField)

	for name, field := range schema {
		virt[name] = field
	}

	if err := mgr.RunInTransaction(func(tx *pg.Tx) error {
		if err := manager.Lock(mgr.ForClassByIDQ(class, o)); err != nil {
			if err == pg.ErrNoRows {
				return api.NewNotFoundError(o)
			}

			return err
		}
		o.Snapshot(o, virt)
		o.ID = 1
		o.Snapshot(o, virt)
		fmt.Println("changed", o.Changes())
		fmt.Println("changed virt", o.ChangesVirtual())

		// Update
		// Validate size
		// Snapshot and compute changes
		return nil
	}); err != nil {
		return err
	}

	return api.NewPermissionDeniedError()
}

func (ctr *Controller) DataObjectDelete(c echo.Context) error {
	o := detailDataObject(c)
	if o == nil {
		return api.NewNotFoundError(o)
	}

	class := c.Get(contextClassKey).(*models.Class)
	mgr := ctr.q.NewDataObjectManager(c)

	return api.SimpleDelete(c, mgr, mgr.ForClassByIDQ(class, o), o)
}

func (ctr *Controller) dataObjectDeleteHook(c database.DBContext, db orm.DB, i interface{}) error {
	o := i.(*models.DataObject)
	sizeDiff := 0

	for k, v := range o.Files.Map {
		fname := o.Data.Map[k]
		util.Must(
			ctr.fs.Default().Delete(context.Background(), settings.BucketData, fname.String),
		)

		if d, e := models.ValueFromString(models.FieldIntegerType, v.String); e == nil {
			sizeDiff += d.(int)
		}
	}

	if sizeDiff != 0 {
		sub := c.Get(contextSubscriptionKey).(*models.Subscription)
		c.Get(contextAdminLimitKey).(*models.AdminLimit).StorageLimit(sub)

		return ctr.updateInstanceIndicatorValue(c, db, models.InstanceIndicatorTypeStorageSize, -sizeDiff)
	}

	return nil
}

func (ctr *Controller) uploadDataObjectFile(ctx context.Context, db orm.DB, instance *models.Instance, class *models.Class, fh *multipart.FileHeader) error { // nolint - ignore that it is unused for now
	key := fmt.Sprintf("%s/%d/%s%s",
		instance.StoragePrefix,
		class.ID,
		util.GenerateHexKey(),
		util.Truncate(filepath.Ext(fh.Filename), 16),
	)

	f, err := fh.Open()
	if err != nil {
		return err
	}

	defer f.Close()

	return storage.SafeUpload(ctx, ctr.fs.Default(), ctr.db, db, settings.BucketData, key, f)
}
