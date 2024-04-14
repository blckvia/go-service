package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetGoodsId(c *gin.Context) (int, error) {
	goodsIDStr := c.Param("id")
	if goodsIDStr == "" {
		newErrorResponse(c, http.StatusBadRequest, "Goods ID is required")
		return 0, errors.New("goods id not found")
	}

	goodsID, err := strconv.Atoi(goodsIDStr)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "Invalid Goods ID format")
		return 0, errors.New("goods id not found")
	}

	return goodsID, nil
}

func GetProjectId(c *gin.Context) (int, error) {
	projectIDStr := c.Param("project_id")
	if projectIDStr == "" {
		newErrorResponse(c, http.StatusBadRequest, "Project ID is required")
		return 0, errors.New("project id not found")
	}

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "Invalid Project ID format")
		return 0, errors.New("project id not found")
	}

	return projectID, nil
}
