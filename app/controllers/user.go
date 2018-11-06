package controllers

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/query"
	"github.com/Syncano/orion/app/serializers"
	"github.com/Syncano/orion/app/validators"
	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/util"
)

const (
	contextUserKey      = "user"
	contextUserClassKey = "user_class"
)

func detailUserObject(c echo.Context) *models.User {
	o := &models.User{}
	v, ok := api.IntParam(c, "user_id", o)
	if !ok {
		return nil
	}
	o.ID = v
	return o
}

// UserContext ...
func UserContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		o := detailUserObject(c)
		if o == nil {
			return api.NewNotFoundError(o)
		}
		if query.NewUserManager(c).OneByID(o) != nil {
			return api.NewNotFoundError(o)
		}

		c.Set(contextUserKey, o)
		return next(c)
	}
}

// UserClassContext ...
func UserClassContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		o := &models.Class{Name: models.UserClassName}
		if query.NewClassManager(c).OneByName(o) != nil {
			return api.NewNotFoundError(o)
		}

		c.Set(contextUserClassKey, o)
		return next(c)
	}
}

// UserCreate ...
func UserCreate(c echo.Context) error {
	// TODO: CORE-2469 create
	return api.NewPermissionDeniedError()
}

// UserList ...
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

	if util.NonEmptyString(c.QueryParam(orderByQuery), "id") != "id" {
		paginator = &PaginatorOrderedDB{Query: q, OrderFields: class.OrderFields()}
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

// UserRetrieve ...
func UserRetrieve(c echo.Context) error {
	o := detailUserObject(c)
	if o == nil {
		return api.NewNotFoundError(o)
	}
	class := c.Get(contextUserClassKey).(*models.Class)

	if query.NewUserManager(c).ByIDQ(class, o).Select() != nil {
		return api.NewNotFoundError(o)
	}

	serializer := serializers.UserSerializer{Class: class}
	return api.Render(c, http.StatusOK, serializer.Response(o))
}

// UserUpdate ...
func UserUpdate(c echo.Context) error {
	// TODO: CORE-2469 updates
	return api.NewPermissionDeniedError()
}

// UserAuth ...
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

// UserDelete ...
func UserDelete(c echo.Context) error {
	// TODO: CORE-2469 deletion
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

// UserSchemaRetrieve ...
func UserSchemaRetrieve(c echo.Context) error {
	class := c.Get(contextUserClassKey).(*models.Class)
	var err error
	if class.ObjectsCount, err = query.NewUserManager(c).CountEstimate(); err != nil {
		return err
	}
	return api.Render(c, http.StatusOK, serializers.UserClassSerializer{}.Response(class))
}

// UserSchemaUpdate ...
func UserSchemaUpdate(c echo.Context) error {
	// TODO: CORE-2433 updates
	return api.NewPermissionDeniedError()
}

// UserMeRetrieve ...
func UserMeRetrieve(c echo.Context) error {
	class := c.Get(contextUserClassKey).(*models.Class)
	user := c.Get(settings.ContextUserKey).(*models.User)
	if err := query.NewUserManager(c).FetchData(class, user); err != nil {
		return err
	}
	serializer := serializers.UserSerializer{Class: class}
	return api.Render(c, http.StatusOK, serializer.ResponseWithGroup(user))
}

// UserMeUpdate ...
func UserMeUpdate(c echo.Context) error {
	// TODO: CORE-2469 updates
	return api.NewPermissionDeniedError()
}

// UserResetKey ...
func UserResetKey(c echo.Context) error {
	// TODO: CORE-2469 updates
	return api.NewPermissionDeniedError()
}

// UsersInGroupCreate ...
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

// UsersInGroupList ...
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

// UsersInGroupRetrieve ...
func UsersInGroupRetrieve(c echo.Context) error {
	o := detailUserObject(c)
	if o == nil {
		return api.NewNotFoundError(o)
	}
	class := c.Get(contextUserClassKey).(*models.Class)
	group := c.Get(contextUserGroupKey).(*models.UserGroup)

	if query.NewUserManager(c).ForGroupByIDQ(class, group, o).Select() != nil {
		return api.NewNotFoundError(o)
	}

	serializer := serializers.UserSerializer{Class: class}
	return api.Render(c, http.StatusOK, serializer.Response(o))
}

// UsersInGroupDelete ...
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
