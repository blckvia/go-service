package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-service/internal/models"
	"go-service/internal/repository"
)

func (h *Handler) createGoods(c *gin.Context) {
	projectID, err := GetProjectId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	var input models.Goods
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if input.Name == "" {
		newErrorResponse(c, http.StatusBadRequest, "name is required")
		return
	}
	if input.Description == "" {
		input.Description = input.Name
	}

	id, err := h.services.Goods.Create(projectID, input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	response := &models.Goods{
		ID:          id,
		ProjectID:   projectID,
		Name:        input.Name,
		Description: input.Description,
		Priority:    input.Priority,
		Removed:     input.Removed,
		CreatedAt:   input.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) getAllGoods(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	goods, err := h.services.Goods.GetAll(limit, offset)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, goods)
}

func (h *Handler) getOne(c *gin.Context) {
	goodsID, err := GetGoodsId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	projectID, err := GetProjectId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	goods, err := h.services.Goods.GetOne(goodsID, projectID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, goods)
}

func (h *Handler) updateGoods(c *gin.Context) {
	goodsID, err := GetGoodsId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	projectID, err := GetProjectId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	var input models.UpdateGoods
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.services.Goods.Update(goodsID, projectID, input); err != nil {
		fmt.Println(errors.Is(err, repository.ErrNotFound))
		if errors.Is(err, repository.ErrNotFound) {
			newDetailedErrorResponse(c, http.StatusNotFound, 3, "errors.good.NotFound", "record not found")
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	updatedGoods, err := h.services.Goods.GetOne(goodsID, projectID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, updatedGoods)
}

func (h *Handler) deleteGoods(c *gin.Context) {
	goodsID, err := GetGoodsId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	projectID, err := GetProjectId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	err = h.services.Goods.Delete(goodsID, projectID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{"ok"})
}
