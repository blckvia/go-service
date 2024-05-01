package handler

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"

	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	_ "go-service/docs"

	"go-service/internal/service"
)

type Handler struct {
	services *service.Service
	tracer   trace.Tracer
}

func New(services *service.Service, tracer trace.Tracer) *Handler {
	return &Handler{
		services: services,
		tracer:   tracer,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.Use(otelgin.Middleware("go-service"))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	debugMode := viper.GetBool("debug")
	if debugMode {
		pprof.Register(router, "/debug/pprof")
	}

	// TODO: /metrics or config telemetry /goods_metrics ?
	// TODO: взять к6 утилиту и написать нагрузку для плофайлера
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
