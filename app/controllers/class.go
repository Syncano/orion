package controllers

import (
	"net/http"

	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/query"
	"github.com/Syncano/orion/app/serializers"
)

const contextClassKey = "class"

func (ctr *Controller) ClassContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		o := &models.Class{Name: c.Param("class_name")}
		if err := ctr.q.NewClassManager(c).OneByName(o); err != nil || !o.Visible {
			if err == pg.ErrNoRows {
				return api.NewNotFoundError(o)
			}

			return err
		}

		c.Set(contextClassKey, o)

		return next(c)
	}
}

func (ctr *Controller) ClassCreate(c echo.Context) error {
	// TODO: #13 Class create
	return api.NewPermissionDeniedError()
}

func (ctr *Controller) ClassList(c echo.Context) error {
	var o []*models.Class

	paginator := &PaginatorDB{Query: ctr.q.NewClassManager(c).WithAccessQ(&o)}
	cursor := paginator.CreateCursor(c, true)

	r, err := Paginate(c, cursor, (*models.Class)(nil), serializers.ClassSerializer{}, paginator)
	if err != nil {
		return err
	}

	return api.Render(c, http.StatusOK, serializers.CreatePage(c, r, nil))
}

func detailClass(c echo.Context) *models.Class {
	return &models.Class{Name: c.Param("class_name")}
}

func (ctr *Controller) ClassRetrieve(c echo.Context) error {
	o := detailClass(c)

	if err := ctr.q.NewClassManager(c).WithAccessByNameQ(o).Select(); err != nil {
		if err == pg.ErrNoRows {
			return api.NewNotFoundError(o)
		}

		return err
	}

	return api.Render(c, http.StatusOK, serializers.ClassSerializer{}.Response(o))
}

func (ctr *Controller) ClassUpdate(c echo.Context) error {
	// TODO: #8 Class updates/deletes
	mgr := ctr.q.NewClassManager(c)
	o := detailClass(c)

	if err := mgr.RunInTransaction(func(*pg.Tx) error {
		err := query.Lock(mgr.WithAccessByNameQ(o))
		if err == pg.ErrNoRows {
			return api.NewNotFoundError(o)
		}

		return err
	}); err != nil {
		return err
	}

	return api.NewPermissionDeniedError()
}

func (ctr *Controller) ClassDelete(c echo.Context) error {
	// TODO: #8 Class updates/deletes
	// index cleanup, DO cascade!
	//
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
