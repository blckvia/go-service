package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-service/internal/models"
	"go-service/internal/repository"
)

// @Summary Create item
// @Tags Goods
// @Description Create one item of goods
// @ID create-goods
// @Accept  json
// @Produce  json
// @Param input body models.UpdateGoods true "goods info"
// @Param project_id path int true "project_id"
// @Success 200 {object} models.Goods
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/goods/{project_id} [post]
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

// @Summary Get list of goods
// @Tags Goods
// @Description Get list of goods
// @ID get-goods
// @Accept  json
// @Produce  json
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {array} models.Goods
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/goods/list [get]
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

// @Summary Get one item
// @Tags Goods
// @Description Get one item of goods
// @ID get-item
// @Accept  json
// @Produce  json
// @Param project_id path int true "project_id"
// @Param id path int true "id"
// @Success 200 {object} models.Goods
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/goods/{project_id}/{id} [get]
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

// @Summary Update item
// @Tags Goods
// @Description Update one item of goods
// @ID update-item
// @Accept  json
// @Produce  json
// @Param input body models.UpdateGoods true "goods info"
// @Param project_id path int true "project_id"
// @Param id path int true "id"
// @Success 200 {object} models.Goods
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/goods/{project_id}/{id} [patch]
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

// @Summary Delete item
// @Tags Goods
// @Description Delete one item of goods
// @ID delete-item
// @Accept  json
// @Produce  json
// @Param project_id path int true "project_id"
// @Param id path int true "id"
// @Success 200
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/goods/{project_id}/{id} [delete]
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

	if err := h.services.Goods.Delete(goodsID, projectID); err != nil {
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

// @Summary Reprioritize item
// @Tags Goods
// @Description Reprioritize one item of goods
// @ID reprioritize-item
// @Accept  json
// @Produce  json
// @Param priority query int true "priority"
// @Param project_id path int true "project_id"
// @Param id path int true "id"
// @Success 200 {object} models.Goods
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/goods/prioritize/{project_id}/{id} [patch]
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

	err = h.services.Goods.Reprioritize(goodsID, projectID, priority)
	if err != nil {
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
