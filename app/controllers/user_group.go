package controllers

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/query"
	"github.com/Syncano/orion/app/serializers"
	"github.com/Syncano/orion/app/validators"
)

const contextUserGroupKey = "user_group"

func detailUserGroup(c echo.Context) *models.UserGroup {
	o := &models.UserGroup{}

	v, ok := api.IntParam(c, "group_id")
	if !ok {
		return nil
	}

	o.ID = v

	return o
}

func UserGroupContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		o := detailUserGroup(c)
		if o == nil {
			return api.NewNotFoundError(o)
		}

		if query.NewUserGroupManager(c).OneByID(o) != nil {
			return api.NewNotFoundError(o)
		}

		c.Set(contextUserGroupKey, o)

		return next(c)
	}
}

func UserGroupCreate(c echo.Context) error {
	// TODO: #15 Users updates/deletes
	return api.NewPermissionDeniedError()
}

func UserGroupList(c echo.Context) error {
	var o []*models.UserGroup

	props := make(map[string]interface{})
	paginator := &PaginatorDB{Query: query.NewUserGroupManager(c).Q(&o)}
	cursor := paginator.CreateCursor(c, true)

	r, err := Paginate(c, cursor, (*models.UserGroup)(nil), serializers.UserGroupSerializer{}, paginator)
	if err != nil {
		return err
	}

	return api.Render(c, http.StatusOK, serializers.CreatePage(c, r, props))
}

func UserGroupRetrieve(c echo.Context) error {
	o := detailUserGroup(c)
	if o == nil {
		return api.NewNotFoundError(o)
	}

	if query.NewUserGroupManager(c).ByIDQ(o).Select() != nil {
		return api.NewNotFoundError(o)
	}

	return api.Render(c, http.StatusOK, serializers.UserGroupSerializer{}.Response(o))
}

func UserGroupUpdate(c echo.Context) error {
	// TODO: #15 Users updates/deletes
	return api.NewPermissionDeniedError()
}

func UserGroupDelete(c echo.Context) error {
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

func GroupsInUserCreate(c echo.Context) error {
	user := c.Get(contextUserKey).(*models.User)
	group := &models.UserGroup{}
	o := &models.UserMembership{UserID: user.ID}
	mgr := query.NewUserMembershipManager(c)
	v := &validators.GroupInUserForm{
		GroupQ:      query.NewUserGroupManager(c).Q(group),
		MembershipQ: mgr.ForUserQ(user, (*models.UserMembership)(nil)),
	}

	if err := api.BindValidateAndExec(c, v, func() error {
		v.Bind(o)
		return mgr.Insert(o)
	}); err != nil {
		return err
	}

	return api.Render(c, http.StatusCreated, serializers.UserGroupSerializer{}.Response(group))
}

func GroupsInUserList(c echo.Context) error {
	var o []*models.UserGroup

	props := make(map[string]interface{})
	user := c.Get(contextUserKey).(*models.User)
	paginator := &PaginatorDB{Query: query.NewUserGroupManager(c).ForUserQ(user, &o)}
	cursor := paginator.CreateCursor(c, true)

	r, err := Paginate(c, cursor, (*models.UserGroup)(nil), serializers.UserGroupSerializer{}, paginator)
	if err != nil {
		return err
	}

	return api.Render(c, http.StatusOK, serializers.CreatePage(c, r, props))
}

func GroupsInUserRetrieve(c echo.Context) error {
	o := detailUserGroup(c)
	if o == nil {
		return api.NewNotFoundError(o)
	}

	user := c.Get(contextUserKey).(*models.User)
	if query.NewUserGroupManager(c).ForUserByIDQ(user, o).Select() != nil {
		return api.NewNotFoundError(o)
	}

	return api.Render(c, http.StatusOK, serializers.UserGroupSerializer{}.Response(o))
}

func GroupsInUserDelete(c echo.Context) error {
	group := detailUserGroup(c)
	if group == nil {
		return api.NewNotFoundError(group)
	}

	mgr := query.NewUserMembershipManager(c)
	user := c.Get(contextUserKey).(*models.User)
	o := &models.UserMembership{UserID: user.ID, GroupID: group.ID}

	return api.SimpleDelete(c, mgr, mgr.ForUserAndGroupQ(o), o)
}
