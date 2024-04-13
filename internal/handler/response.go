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

type detailedErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
	Detail  string `json:"detail"`
}

func newErrorResponse(c *gin.Context, statusCode int, message string) {
	logger.GetLogger().Error(message)
	c.AbortWithStatusJSON(statusCode, errorResponse{message})
}

func newDetailedErrorResponse(c *gin.Context, statusCode, errorCode int, message, detail string) {
	logger.GetLogger().Error(message)
	c.AbortWithStatusJSON(statusCode, detailedErrorResponse{
		Code:    errorCode,
		Message: message,
		Status:  statusCode,
		Detail:  detail,
	})
}
