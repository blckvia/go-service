package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

const goodsCtx = "id"

func GetGoodsId(c *gin.Context) (int, error) {
	id, ok := c.Get(goodsCtx)
	if !ok {
		newErrorResponse(c, http.StatusNotFound, "goods id not found")
		return 0, errors.New("goods id not found")
	}

	idInt, ok := id.(int)
	if !ok {
		newErrorResponse(c, http.StatusNotFound, "goods id not found")
		return 0, errors.New("goods id not found")
	}

	return idInt, nil
}

func GetProjectId(c *gin.Context) (int, error) {
	id, ok := c.Get(goodsCtx)
	if !ok {
		newErrorResponse(c, http.StatusNotFound, "project id not found")
		return 0, errors.New("project id not found")
	}

	idInt, ok := id.(int)
	if !ok {
		newErrorResponse(c, http.StatusNotFound, "project id not found")
		return 0, errors.New("project id not found")
	}

	return idInt, nil
}
