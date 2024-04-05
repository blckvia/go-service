package handler

import (
	"github.com/gin-gonic/gin"

	"go-service/pkg/logger"
)

type errorResponse struct {
	Message string `json:"message"`
}

type statusResponse struct {
	Status string `json:"status"`
}

func newErrorResponse(c *gin.Context, statusCode int, message string) {
	logger.GetLogger().Error(message)
	c.AbortWithStatusJSON(statusCode, errorResponse{message})
}