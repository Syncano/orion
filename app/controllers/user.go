package controllers

import (
	"net/http"

	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/serializers"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/orion/app/validators"
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

func (ctr *Controller) UserContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		o := detailUserObject(c)
		if o == nil {
			return api.NewNotFoundError(o)
		}

		if err := ctr.q.NewUserManager(c).OneByID(o); err != nil {
			if err == pg.ErrNoRows {
				return api.NewNotFoundError(o)
			}

			return err
		}

		c.Set(contextUserKey, o)

		return next(c)
	}
}

func (ctr *Controller) UserClassContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		o := &models.Class{Name: models.UserClassName}
		if err := ctr.q.NewClassManager(c).OneByName(o); err != nil {
			if err == pg.ErrNoRows {
				return api.NewNotFoundError(o)
			}

			return err
		}

		c.Set(contextUserClassKey, o)

		return next(c)
	}
}

func (ctr *Controller) UserCreate(c echo.Context) error {
	// TODO: #15 Users updates/deletes
	return api.NewPermissionDeniedError()
}

func (ctr *Controller) UserList(c echo.Context) error {
	var o []*models.User

	mgr := ctr.q.NewUserManager(c)
	props := make(map[string]interface{})
	class := c.Get(contextUserClassKey).(*models.Class)

	// Prepare query.
	q := mgr.Q(class, &o)

	if _, e := c.QueryParams()["query"]; e {
		var err error
		q, err = NewDataObjectQuery(class.FilterFields()).Parse(ctr.q, c, q)

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

func (ctr *Controller) UserRetrieve(c echo.Context) error {
	o := detailUserObject(c)
	if o == nil {
		return api.NewNotFoundError(o)
	}

	class := c.Get(contextUserClassKey).(*models.Class)

	if err := ctr.q.NewUserManager(c).ByIDQ(class, o).Select(); err != nil {
		if err == pg.ErrNoRows {
			return api.NewNotFoundError(o)
		}

		return err
	}

	serializer := serializers.UserSerializer{Class: class}

	return api.Render(c, http.StatusOK, serializer.Response(o))
}

func (ctr *Controller) UserUpdate(c echo.Context) error {
	// TODO: #15 Users updates/deletes
	return api.NewPermissionDeniedError()
}

func (ctr *Controller) UserAuth(c echo.Context) error {
	form := &validators.UserAuthForm{}
	if err := api.BindAndValidate(c, form); err != nil {
		return err
	}

	o := &models.User{Username: form.Username}
	mgr := ctr.q.NewUserManager(c)
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

func (ctr *Controller) UserDelete(c echo.Context) error {
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

func (ctr *Controller) UserSchemaRetrieve(c echo.Context) error {
	class := c.Get(contextUserClassKey).(*models.Class)

	var err error
	if class.ObjectsCount, err = ctr.q.NewUserManager(c).CountEstimate(); err != nil {
		return err
	}

	return api.Render(c, http.StatusOK, serializers.UserClassSerializer{}.Response(class))
}

func (ctr *Controller) UserSchemaUpdate(c echo.Context) error {
	// TODO: #8 Class updates/deletes
	return api.NewPermissionDeniedError()
}

func (ctr *Controller) UserMeRetrieve(c echo.Context) error {
	class := c.Get(contextUserClassKey).(*models.Class)
	user := c.Get(settings.ContextUserKey).(*models.User)

	if err := ctr.q.NewUserManager(c).FetchData(class, user); err != nil {
		return err
	}

	serializer := serializers.UserSerializer{Class: class}

	return api.Render(c, http.StatusOK, serializer.ResponseWithGroup(user))
}

func (ctr *Controller) UserMeUpdate(c echo.Context) error {
	// TODO: #15 Users updates/deletes
	return api.NewPermissionDeniedError()
}

func (ctr *Controller) UserResetKey(c echo.Context) error {
	// TODO: #15 Users updates/deletes
	return api.NewPermissionDeniedError()
}

func (ctr *Controller) UsersInGroupCreate(c echo.Context) error {
	class := c.Get(contextUserClassKey).(*models.Class)
	group := c.Get(contextUserGroupKey).(*models.UserGroup)
	user := &models.User{}
	o := &models.UserMembership{GroupID: group.ID}
	mgr := ctr.q.NewUserMembershipManager(c)

	v := &validators.UserInGroupForm{
		UserQ:       ctr.q.NewUserManager(c).Q(class, user),
		MembershipQ: mgr.ForGroupQ(group, (*models.UserMembership)(nil)),
	}

	if err := api.BindValidateAndExec(c, v, func() error {
		v.Bind(o)
		return mgr.InsertContext(c.Request().Context(), o)
	}); err != nil {
		return err
	}

	serializer := serializers.UserSerializer{Class: class}

	return api.Render(c, http.StatusCreated, serializer.Response(user))
}

func (ctr *Controller) UsersInGroupList(c echo.Context) error {
	var o []*models.User

	props := make(map[string]interface{})
	class := c.Get(contextUserClassKey).(*models.Class)
	group := c.Get(contextUserGroupKey).(*models.UserGroup)
	paginator := &PaginatorDB{Query: ctr.q.NewUserManager(c).ForGroupQ(class, group, &o)}
	cursor := paginator.CreateCursor(c, true)
	serializer := serializers.UserSerializer{Class: class}

	r, err := Paginate(c, cursor, (*models.User)(nil), serializer, paginator)
	if err != nil {
		return err
	}

	return api.Render(c, http.StatusOK, serializers.CreatePage(c, r, props))
}

func (ctr *Controller) UsersInGroupRetrieve(c echo.Context) error {
	o := detailUserObject(c)
	if o == nil {
		return api.NewNotFoundError(o)
	}

	class := c.Get(contextUserClassKey).(*models.Class)
	group := c.Get(contextUserGroupKey).(*models.UserGroup)

	if err := ctr.q.NewUserManager(c).ForGroupByIDQ(class, group, o).Select(); err != nil {
		if err == pg.ErrNoRows {
			return api.NewNotFoundError(o)
		}

		return err
	}

	serializer := serializers.UserSerializer{Class: class}

	return api.Render(c, http.StatusOK, serializer.Response(o))
}

func (ctr *Controller) UsersInGroupDelete(c echo.Context) error {
	user := detailUserObject(c)
	if user == nil {
		return api.NewNotFoundError(user)
	}

	mgr := ctr.q.NewUserMembershipManager(c)
	group := c.Get(contextUserGroupKey).(*models.UserGroup)
	o := &models.UserMembership{UserID: user.ID, GroupID: group.ID}

	return api.SimpleDelete(c, mgr, mgr.ForUserAndGroupQ(o), o)
}
