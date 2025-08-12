package handler

import "github.com/gin-gonic/gin"

type Handler struct {
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	auth := router.Group("/auth")
	{
		auth.POST("/sign-in", h.signIn)
		auth.POST("/sign-out", h.signOut)
	}

	api := router.Group("/api")
	{
		groups := api.Group("/groups")
		{
			groups.GET("/", h.getAllGroups)
			groups.POST("/", h.createGroup)
			groups.GET("/:id", h.getGroupById)
			groups.PATCH("/:id", h.updateGroup)
			groups.DELETE("/:id", h.deleteGroup)
			// groups.GET("/:id/users")
			// groups.POST("/:id/users")

			// policies := groups.Group("/:id/policy")
			// {
			// 	policies.GET("/")
			// 	policies.PATCH("/")
			// }
		}

		users := api.Group("/users")
		{
			users.GET("/", h.getAllUsers)
			users.POST("/", h.createUser)
			users.GET("/:id", h.getUserById)
			users.PATCH("/:id", h.updateUser)
			users.DELETE("/:id", h.deleteUser)
		}

		// whitelist := api.Group("/whitelist")
		// {
		// 	whitelist.GET("/")
		// 	whitelist.POST("/")
		// 	whitelist.PATCH("/:id")
		// 	whitelist.DELETE("/:id")
		// }
	}
	return router
}