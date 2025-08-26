package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/rinat0880/classOS_backend/pkg/service"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	auth := router.Group("/auth")
	{
		auth.POST("/sign-up", h.signUp)
		auth.POST("/sign-in", h.signIn)
	}

	api := router.Group("/api", h.userIdentity, h.adminOnly)
	{
		groups := api.Group("/groups")
		{
			groups.GET("/", h.getAllGroups)
			groups.POST("/", h.createGroup)
			groups.GET("/:id", h.getGroupById)
			groups.PATCH("/:id", h.updateGroup)
			groups.DELETE("/:id", h.deleteGroup)

			// policies := groups.Group("/:id/policy")
			// {
			// 	policies.GET("/")
			// 	policies.PATCH("/")
			// }
			users := groups.Group(":id/users")
			{
				users.GET("/", h.getAllUsers)
				users.POST("/", h.createUser)
			}
		}

		users := api.Group("/users")
		{
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
