package controllers

import (
	"net/http"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/serializers"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/pkg-go/v2/database/manager"
)

func (ctr *Controller) InstanceContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		o := &models.Instance{Name: c.Param("instance_name")}
		if err := ctr.q.NewInstanceManager(c).OneByName(o); err != nil {
			if err == pg.ErrNoRows {
				return api.NewNotFoundError(o)
			}

			return err
		}

		if o.Location != settings.Common.Location {
			return api.NewBadRequestError("Instance was created in different location. Use relevant API endpoint.")
		}

		// Get Instance owner and check last access time.
		var owner *models.Admin

		if a := c.Get(settings.ContextAdminKey); a != nil {
			adm := a.(*models.Admin)
			if adm.ID == o.OwnerID {
				owner = adm
			}
		}

		adminMgr := ctr.q.NewAdminManager(c)

		if owner == nil {
			owner = &models.Admin{ID: o.OwnerID}
			if err := adminMgr.OneByID(owner); err != nil {
				if err == pg.ErrNoRows {
					return api.NewNotFoundError(o)
				}

				return err
			}
		}

		if owner.LastAccess.IsNull() || time.Since(owner.LastAccess.Time) > 12*time.Hour {
			owner.LastAccess.Set(time.Now())                    // nolint: errcheck
			owner.NoticedAt.Set(nil)                            // nolint: errcheck
			adminMgr.Update(owner, "last_access", "noticed_at") // nolint: errcheck
		}

		c.Set(settings.ContextInstanceKey, o)
		c.Set(settings.ContextInstanceOwnerKey, owner)
		c.Set(settings.ContextSchemaKey, o.SchemaName)

		return next(c)
	}
}

func (ctr *Controller) InstanceAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		o := c.Get(settings.ContextInstanceKey).(*models.Instance)
		perm := false

		if a := c.Get(settings.ContextAdminKey); a != nil {
			adm := a.(*models.Admin)
			air := &models.AdminInstanceRole{InstanceID: o.ID, AdminID: adm.ID}

			if adm.ID == o.OwnerID || adm.IsStaff || ctr.q.NewAdminInstanceRoleManager(c).OneByInstanceAndAdmin(air) == nil {
				perm = true
			}
		} else if a := c.Get(settings.ContextAPIKeyKey); a != nil {
			perm = a.(*models.APIKey).InstanceID == o.ID
		}

		if !perm {
			return api.NewPermissionDeniedError()
		}

		return next(c)
	}
}

func (ctr *Controller) InstanceCreate(c echo.Context) error {
	// TODO: #12 Instance create
	return api.NewPermissionDeniedError()
}

func (ctr *Controller) InstanceList(c echo.Context) error {
	var o []*models.Instance

	paginator := &PaginatorDB{Query: ctr.q.NewInstanceManager(c).WithAccessQ(&o)}
	cursor := paginator.CreateCursor(c, true)

	r, err := Paginate(c, cursor, (*models.Instance)(nil), serializers.InstanceSerializer{}, paginator)
	if err != nil {
		return err
	}

	return api.Render(c, http.StatusOK, serializers.CreatePage(c, r, nil))
}

func detailInstance(c echo.Context) *models.Instance {
	return &models.Instance{Name: c.Param("instance_name")}
}

func (ctr *Controller) InstanceRetrieve(c echo.Context) error {
	o := detailInstance(c)

	if err := ctr.q.NewInstanceManager(c).WithAccessByNameQ(o).Select(); err != nil {
		if err == pg.ErrNoRows {
			return api.NewNotFoundError(o)
		}

		return err
	}

	return api.Render(c, http.StatusOK, serializers.InstanceSerializer{}.Response(o))
}

func (ctr *Controller) InstanceUpdate(c echo.Context) error {
	// TODO: #9 Instance updates/deletes
	mgr := ctr.q.NewInstanceManager(c)
	o := detailInstance(c)

	if err := mgr.RunInTransaction(func(*pg.Tx) error {
		err := manager.Lock(mgr.WithAccessByNameQ(o))
		if err == pg.ErrNoRows {
			return api.NewNotFoundError(o)
		}

		return err
	}); err != nil {
		return err
	}

	return api.NewPermissionDeniedError()
}

func (ctr *Controller) InstanceDelete(c echo.Context) error {
	// TODO: #9 Instance updates/deletes
	// user := detailUserObject(c)
	// if user == nil {
	// 	return api.NewNotFoundError(user)
	// }
	//
	// mgr := query.NewUserMembershipManager(c)
	// group := c.Get(contextUserGroupKey).(*models.UserGroup)
	// o := &models.UserMembership{UserID: user.ID, GroupID: group.ID}
	//
	// return api.SimpleDelete(c, mgr, mgr.ForUserAndGroupQ(o), o)
	return api.NewPermissionDeniedError()
}
