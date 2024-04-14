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
			projects.GET("/:project_id", h.getProject)
			projects.PATCH("/:project_id", h.updateProject)
			projects.DELETE("/:project_id", h.deleteProject)
		}

		goods := api.Group("/goods")
		{
			goods.GET("/list", h.getAllGoods)
			goods.PATCH("/prioritize/:project_id/:id", h.reprioritize)
			goods.POST("/:project_id", h.createGoods)
			goods.PATCH("/:project_id/:id", h.updateGoods)
			goods.DELETE("/:project_id/:id", h.deleteGoods)
			goods.GET("/:project_id/:id", h.getOne)
		}
	}

	return router
}
