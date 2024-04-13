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
		//projects := api.Group("/projects")
		//{
		//	projects.POST("/", h.createProject)
		//	projects.GET("/", h.getAllProjects)
		//	projects.GET("/:id", h.getProject)
		//	projects.PATCH("/:id", h.updateProject)
		//	projects.DELETE("/:id", h.deleteProject)
		//}

		goods := api.Group("/goods")
		{
			goods.GET("/list", h.getAllGoods)
			goods.POST("/:project_id", h.createGoods)
			goods.PATCH("/:project_id/:id", h.updateGoods)
			goods.DELETE("/:id", h.deleteGoods)
			goods.GET("/:project_id/:id", h.getOne)
		}
	}

	return router
}
