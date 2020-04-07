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

func ClassContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		o := &models.Class{Name: c.Param("class_name")}
		if query.NewClassManager(c).OneByName(o) != nil || !o.Visible {
			return api.NewNotFoundError(o)
		}

		c.Set(contextClassKey, o)

		return next(c)
	}
}

func ClassCreate(c echo.Context) error {
	// TODO: #13 Class create
	return api.NewPermissionDeniedError()
}

func ClassList(c echo.Context) error {
	var o []*models.Class

	paginator := &PaginatorDB{Query: query.NewClassManager(c).WithAccessQ(&o)}
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

func ClassRetrieve(c echo.Context) error {
	o := detailClass(c)

	if err := query.NewClassManager(c).WithAccessByNameQ(o).Select(); err != nil {
		if err == pg.ErrNoRows {
			return api.NewNotFoundError(o)
		}

		return err
	}

	return api.Render(c, http.StatusOK, serializers.ClassSerializer{}.Response(o))
}

func ClassUpdate(c echo.Context) error {
	// TODO: #8 Class updates/deletes
	mgr := query.NewClassManager(c)
	o := detailClass(c)

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

func ClassDelete(c echo.Context) error {
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
