package handler

import (
	"errors"
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

	id, err := h.services.Goods.Create(c, projectID, input)
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
	goods, err := h.services.Goods.GetAll(c, limit, offset)
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

	goods, err := h.services.Goods.GetOne(c, goodsID, projectID)
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

	if err := h.services.Goods.Update(c, goodsID, projectID, input); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			newDetailedErrorResponse(c, http.StatusNotFound, 3, "errors.good.NotFound", "record not found")
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	updatedGoods, err := h.services.Goods.GetOne(c, goodsID, projectID)
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

	err = h.services.Goods.Delete(c, goodsID, projectID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.services.Goods.Delete(c, goodsID, projectID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			newDetailedErrorResponse(c, http.StatusNotFound, 3, "errors.good.NotFound", "record not found")
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	updatedGoods, err := h.services.Goods.GetOne(c, goodsID, projectID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, updatedGoods)
}

func (h *Handler) reprioritize(c *gin.Context) {
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

	priority, err := strconv.Atoi(c.Query("priority"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.services.Goods.Reprioritize(c, goodsID, projectID, priority)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			newDetailedErrorResponse(c, http.StatusNotFound, 3, "errors.good.NotFound", "record not found")
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	updatedGoods, err := h.services.Goods.GetOne(c, goodsID, projectID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, updatedGoods)
}
