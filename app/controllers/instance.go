package controllers

import (
	"net/http"
	"time"

	"github.com/go-pg/pg"
	"github.com/labstack/echo"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/query"
	"github.com/Syncano/orion/app/serializers"
	"github.com/Syncano/orion/pkg/settings"
)

// InstanceContext ...
func InstanceContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		o := &models.Instance{Name: c.Param("instance_name")}
		if query.NewInstanceManager(c).OneByName(o) != nil {
			return api.NewNotFoundError(o)
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
		adminMgr := query.NewAdminManager(c)
		if owner == nil {
			owner = &models.Admin{ID: o.OwnerID}
			if adminMgr.OneByID(owner) != nil {
				return api.NewNotFoundError(o)
			}
		}

		if true || owner.LastAccess.IsNull() || time.Since(owner.LastAccess.Time) > 12*time.Hour {
			owner.LastAccess.Set(time.Now())                    // nolint: errcheck
			owner.NoticedAt.Set(nil)                            // nolint: errcheck
			adminMgr.Update(owner, "last_access", "noticed_at") // nolint: errcheck
		}

		c.Set(settings.ContextInstanceKey, o)
		c.Set(settings.ContextInstanceOwnerKey, owner)
		c.Set(query.ContextSchemaKey, o.SchemaName)
		return next(c)
	}
}

// InstanceAuth ...
func InstanceAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		o := c.Get(settings.ContextInstanceKey).(*models.Instance)

		perm := false
		if a := c.Get(settings.ContextAdminKey); a != nil {
			adm := a.(*models.Admin)
			air := &models.AdminInstanceRole{InstanceID: o.ID, AdminID: adm.ID}

			if adm.ID == o.OwnerID || adm.IsStaff || query.NewAdminInstanceRoleManager(c).OneByInstanceAndAdmin(air) == nil {
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

// InstanceCreate ...
func InstanceCreate(c echo.Context) error {
	// TODO: #12 Instance create
	return api.NewPermissionDeniedError()
}

// InstanceList ...
func InstanceList(c echo.Context) error {
	var o []*models.Instance
	paginator := &PaginatorDB{Query: query.NewInstanceManager(c).WithAccessQ(&o)}
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

// InstanceRetrieve ...
func InstanceRetrieve(c echo.Context) error {
	o := detailInstance(c)

	if query.NewInstanceManager(c).WithAccessByNameQ(o).Select() != nil {
		return api.NewNotFoundError(o)
	}

	return api.Render(c, http.StatusOK, serializers.InstanceSerializer{}.Response(o))
}

// InstanceUpdate ...
func InstanceUpdate(c echo.Context) error {
	// TODO: #9 Instance updates/deletes
	mgr := query.NewInstanceManager(c)
	o := detailInstance(c)

	if err := mgr.RunInTransaction(func(*pg.Tx) error {
		if query.Lock(mgr.WithAccessByNameQ(o)) != nil {
			return api.NewNotFoundError(o)
		}
		return nil
	}); err != nil {
		return err
	}
	return api.NewPermissionDeniedError()
}

// InstanceDelete ...
func InstanceDelete(c echo.Context) error {
	// TODO: #9 Instance updates/deletes
	// user := detailUserObject(c)
	// if user == nil {
	// 	return api.NewNotFoundError(user)
	// }

	// mgr := query.NewUserMembershipManager(c)
	// group := c.Get(contextUserGroupKey).(*models.UserGroup)
	// o := &models.UserMembership{UserID: user.ID, GroupID: group.ID}

	// return api.SimpleDelete(c, mgr, mgr.ForUserAndGroupQ(o), o)
	return api.NewPermissionDeniedError()
}
