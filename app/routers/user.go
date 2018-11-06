package routers

import (
	"github.com/labstack/echo"

	"github.com/Syncano/orion/app/controllers"
)

// UserRegister registers user routes.
func UserRegister(r *echo.Group, m *middlewares) {
	g := r.Group("", m.Add(controllers.UserClassContext).Get()...)

	// List routes.
	// /users/
	g.GET("/", controllers.UserList)
	g.POST("/", controllers.UserCreate)

	// Schema routes.
	// /users/schema/
	d := g.Group("/schema")
	d.GET("/", controllers.UserSchemaRetrieve)
	d.PATCH("/", controllers.UserSchemaUpdate)

	// User Me routes. Require User.
	// /users/me/
	d = g.Group("/me", controllers.RequireUser)
	d.GET("/", controllers.UserMeRetrieve)
	d.PATCH("/", controllers.UserMeUpdate)

	// Auth routes.
	// /users/auth/
	g.POST("/auth/", controllers.UserAuth)

	// Detail routes.
	// /users/:id/
	d = g.Group("/:user_id")
	d.GET("/", controllers.UserRetrieve)
	d.PATCH("/", controllers.UserUpdate)
	d.DELETE("/", controllers.UserDelete)

	// Sub user routes. UserClassContext is no longer needed. Add UserContext instead.
	// /users/:id/reset_key/
	g = r.Group("/:user_id", m.Add(controllers.UserContext).Get()...)
	g.POST("/reset_key/", controllers.UserResetKey)

	// User groups routes.
	// /users/:id/groups/
	g = g.Group("/groups")
	g.GET("/", controllers.GroupsInUserList)
	g.POST("/", controllers.GroupsInUserCreate)

	// /users/:id/groups/:id/
	d = g.Group("/:group_id")
	d.GET("/", controllers.GroupsInUserRetrieve)
	d.DELETE("/", controllers.GroupsInUserDelete)
}

// UserGroupRegister registers user group routes.
func UserGroupRegister(r *echo.Group, m *middlewares) {
	g := r.Group("", m.Get()...)

	// List routes.
	// /groups/
	g.GET("/", controllers.UserGroupList)
	g.POST("/", controllers.UserGroupCreate)

	// Detail routes.
	// /groups/:id/
	d := g.Group("/:group_id")
	d.GET("/", controllers.UserGroupRetrieve)
	d.PATCH("/", controllers.UserGroupUpdate)
	d.DELETE("/", controllers.UserGroupDelete)

	// Sub group routes.
	// /groups/:id/users/
	g = d.Group("/users", controllers.UserClassContext, controllers.UserGroupContext)
	g.GET("/", controllers.UsersInGroupList)
	g.POST("/", controllers.UsersInGroupCreate)

	// /groups/:id/users/:id/
	d = g.Group("/:user_id")
	d.GET("/", controllers.UsersInGroupRetrieve)
	d.DELETE("/", controllers.UsersInGroupDelete)
}
