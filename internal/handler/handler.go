package handler

import (
	"github.com/gin-gonic/gin"

	"go-service/internal/service"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	api := router.Group("/api")
	{
		projects := api.Group("/projects")
		{
			projects.POST("/", h.createProject)
			projects.GET("/", h.getAllProjects)
			projects.GET("/:id", h.getProject)
			projects.PATCH("/:id", h.updateProject)
			projects.DELETE("/:id", h.deleteProject)
		}

		goods := router.Group("/goods")
		{
			goods.POST("/", h.createGoods)
			goods.GET("/", h.getAllGoods)
			goods.GET("/:id", h.getGoodsById)
			goods.PATCH("/:id", h.updateGoods)
			goods.DELETE("/:id", h.deleteGoods)
		}
	}

	return router
}
