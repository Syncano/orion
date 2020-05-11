package routers

import (
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/controllers"
)

// UserRegister registers user routes.
func UserRegister(ctr *controllers.Controller, r *echo.Group, m *middlewares) {
	g := r.Group("", m.Add(ctr.UserClassContext).Get(ctr)...)

	// List routes.
	// /users/
	g.GET("/", ctr.UserList)
	g.POST("/", ctr.UserCreate)

	// Schema routes.
	// /users/schema/
	d := g.Group("/schema")
	d.GET("/", ctr.UserSchemaRetrieve)
	d.PATCH("/", ctr.UserSchemaUpdate)

	// User Me routes. Require User.
	// /users/me/
	d = g.Group("/me", ctr.RequireUser)
	d.GET("/", ctr.UserMeRetrieve)
	d.PATCH("/", ctr.UserMeUpdate)

	// Auth routes.
	// /users/auth/
	g.POST("/auth/", ctr.UserAuth)

	// Detail routes.
	// /users/:id/
	d = g.Group("/:user_id")
	d.GET("/", ctr.UserRetrieve)
	d.PATCH("/", ctr.UserUpdate)
	d.DELETE("/", ctr.UserDelete)

	// Sub user routes. UserClassContext is no longer needed. Add UserContext instead.
	// /users/:id/reset_key/
	g = r.Group("/:user_id", m.Add(ctr.UserContext).Get(ctr)...)
	g.POST("/reset_key/", ctr.UserResetKey)

	// User groups routes.
	// /users/:id/groups/
	g = g.Group("/groups")
	g.GET("/", ctr.GroupsInUserList)
	g.POST("/", ctr.GroupsInUserCreate)

	// /users/:id/groups/:id/
	d = g.Group("/:group_id")
	d.GET("/", ctr.GroupsInUserRetrieve)
	d.DELETE("/", ctr.GroupsInUserDelete)
}

// UserGroupRegister registers user group routes.
func UserGroupRegister(ctr *controllers.Controller, r *echo.Group, m *middlewares) {
	g := r.Group("", m.Get(ctr)...)

	// List routes.
	// /groups/
	g.GET("/", ctr.UserGroupList)
	g.POST("/", ctr.UserGroupCreate)

	// Detail routes.
	// /groups/:id/
	d := g.Group("/:group_id")
	d.GET("/", ctr.UserGroupRetrieve)
	d.PATCH("/", ctr.UserGroupUpdate)
	d.DELETE("/", ctr.UserGroupDelete)

	// Sub group routes.
	// /groups/:id/users/
	g = d.Group("/users", ctr.UserClassContext, ctr.UserGroupContext)
	g.GET("/", ctr.UsersInGroupList)
	g.POST("/", ctr.UsersInGroupCreate)

	// /groups/:id/users/:id/
	d = g.Group("/:user_id")
	d.GET("/", ctr.UsersInGroupRetrieve)
	d.DELETE("/", ctr.UsersInGroupDelete)
}
