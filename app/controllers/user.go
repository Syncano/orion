package controllers

import (
	"net/http"

	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/query"
	"github.com/Syncano/orion/app/serializers"
	"github.com/Syncano/orion/app/validators"
	"github.com/Syncano/orion/pkg/settings"
)

const (
	contextUserKey      = "user"
	contextUserClassKey = "user_class"
)

func detailUserObject(c echo.Context) *models.User {
	o := &models.User{}

	v, ok := api.IntParam(c, "user_id")
	if !ok {
		return nil
	}

	o.ID = v

	return o
}

func UserContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		o := detailUserObject(c)
		if o == nil {
			return api.NewNotFoundError(o)
		}

		if err := query.NewUserManager(c).OneByID(o); err != nil {
			if err == pg.ErrNoRows {
				return api.NewNotFoundError(o)
			}

			return err
		}

		c.Set(contextUserKey, o)

		return next(c)
	}
}

func UserClassContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		o := &models.Class{Name: models.UserClassName}
		if err := query.NewClassManager(c).OneByName(o); err != nil {
			if err == pg.ErrNoRows {
				return api.NewNotFoundError(o)
			}

			return err
		}

		c.Set(contextUserClassKey, o)

		return next(c)
	}
}

func UserCreate(c echo.Context) error {
	// TODO: #15 Users updates/deletes
	return api.NewPermissionDeniedError()
}

func UserList(c echo.Context) error {
	var o []*models.User

	mgr := query.NewUserManager(c)
	props := make(map[string]interface{})
	class := c.Get(contextUserClassKey).(*models.Class)

	// Prepare query.
	q := mgr.Q(class, &o)

	if _, e := c.QueryParams()["query"]; e {
		var err error
		q, err = NewDataObjectQuery(class.FilterFields()).Parse(c, q)

		if err != nil {
			return err
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
	serializer := serializers.UserSerializer{Class: class}

	r, err := Paginate(c, cursor, (*models.User)(nil), serializer, paginator)
	if err != nil {
		return err
	}

	return api.Render(c, http.StatusOK, serializers.CreatePage(c, r, props))
}

func UserRetrieve(c echo.Context) error {
	o := detailUserObject(c)
	if o == nil {
		return api.NewNotFoundError(o)
	}

	class := c.Get(contextUserClassKey).(*models.Class)

	if err := query.NewUserManager(c).ByIDQ(class, o).Select(); err != nil {
		if err == pg.ErrNoRows {
			return api.NewNotFoundError(o)
		}

		return err
	}

	serializer := serializers.UserSerializer{Class: class}

	return api.Render(c, http.StatusOK, serializer.Response(o))
}

func UserUpdate(c echo.Context) error {
	// TODO: #15 Users updates/deletes
	return api.NewPermissionDeniedError()
}

func UserAuth(c echo.Context) error {
	form := &validators.UserAuthForm{}
	if err := api.BindAndValidate(c, form); err != nil {
		return err
	}

	o := &models.User{Username: form.Username}
	mgr := query.NewUserManager(c)
	class := c.Get(contextUserClassKey).(*models.Class)

	if err := mgr.OneByName(o); err != nil {
		return api.NewGenericError(http.StatusUnauthorized, "Invalid username.")
	}

	if !o.CheckPassword(form.Password) {
		return api.NewGenericError(http.StatusUnauthorized, "Invalid password.")
	}

	if err := mgr.FetchData(class, o); err != nil {
		return err
	}

	serializer := serializers.UserSerializer{Class: class}

	return api.Render(c, http.StatusOK, serializer.ResponseWithGroup(o))
}

func UserDelete(c echo.Context) error {
	// TODO: #15 Users updates/deletes
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

func UserSchemaRetrieve(c echo.Context) error {
	class := c.Get(contextUserClassKey).(*models.Class)

	var err error
	if class.ObjectsCount, err = query.NewUserManager(c).CountEstimate(); err != nil {
		return err
	}

	return api.Render(c, http.StatusOK, serializers.UserClassSerializer{}.Response(class))
}

func UserSchemaUpdate(c echo.Context) error {
	// TODO: #8 Class updates/deletes
	return api.NewPermissionDeniedError()
}

func UserMeRetrieve(c echo.Context) error {
	class := c.Get(contextUserClassKey).(*models.Class)
	user := c.Get(settings.ContextUserKey).(*models.User)

	if err := query.NewUserManager(c).FetchData(class, user); err != nil {
		return err
	}

	serializer := serializers.UserSerializer{Class: class}

	return api.Render(c, http.StatusOK, serializer.ResponseWithGroup(user))
}

func UserMeUpdate(c echo.Context) error {
	// TODO: #15 Users updates/deletes
	return api.NewPermissionDeniedError()
}

func UserResetKey(c echo.Context) error {
	// TODO: #15 Users updates/deletes
	return api.NewPermissionDeniedError()
}

func UsersInGroupCreate(c echo.Context) error {
	class := c.Get(contextUserClassKey).(*models.Class)
	group := c.Get(contextUserGroupKey).(*models.UserGroup)
	user := &models.User{}
	o := &models.UserMembership{GroupID: group.ID}
	mgr := query.NewUserMembershipManager(c)

	v := &validators.UserInGroupForm{
		UserQ:       query.NewUserManager(c).Q(class, user),
		MembershipQ: mgr.ForGroupQ(group, (*models.UserMembership)(nil)),
	}

	if err := api.BindValidateAndExec(c, v, func() error {
		v.Bind(o)
		return mgr.Insert(o)
	}); err != nil {
		return err
	}

	serializer := serializers.UserSerializer{Class: class}

	return api.Render(c, http.StatusCreated, serializer.Response(user))
}

func UsersInGroupList(c echo.Context) error {
	var o []*models.User

	props := make(map[string]interface{})
	class := c.Get(contextUserClassKey).(*models.Class)
	group := c.Get(contextUserGroupKey).(*models.UserGroup)
	paginator := &PaginatorDB{Query: query.NewUserManager(c).ForGroupQ(class, group, &o)}
	cursor := paginator.CreateCursor(c, true)
	serializer := serializers.UserSerializer{Class: class}

	r, err := Paginate(c, cursor, (*models.User)(nil), serializer, paginator)
	if err != nil {
		return err
	}

	return api.Render(c, http.StatusOK, serializers.CreatePage(c, r, props))
}

func UsersInGroupRetrieve(c echo.Context) error {
	o := detailUserObject(c)
	if o == nil {
		return api.NewNotFoundError(o)
	}

	class := c.Get(contextUserClassKey).(*models.Class)
	group := c.Get(contextUserGroupKey).(*models.UserGroup)

	if err := query.NewUserManager(c).ForGroupByIDQ(class, group, o).Select(); err != nil {
		if err == pg.ErrNoRows {
			return api.NewNotFoundError(o)
		}

		return err
	}

	serializer := serializers.UserSerializer{Class: class}

	return api.Render(c, http.StatusOK, serializer.Response(o))
}

func UsersInGroupDelete(c echo.Context) error {
	user := detailUserObject(c)
	if user == nil {
		return api.NewNotFoundError(user)
	}

	mgr := query.NewUserMembershipManager(c)
	group := c.Get(contextUserGroupKey).(*models.UserGroup)
	o := &models.UserMembership{UserID: user.ID, GroupID: group.ID}

	return api.SimpleDelete(c, mgr, mgr.ForUserAndGroupQ(o), o)
}
