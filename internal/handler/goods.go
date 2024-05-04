package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"go-service/internal/models"
	"go-service/internal/repository"
	p "go-service/pkg/prometheus"
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

	ctx, span := h.tracer.Start(c.Request.Context(), "createGoods")
	defer span.End()
	span.AddEvent("create goods", trace.WithAttributes(attribute.String("name", input.Name)))

	id, err := h.services.Goods.Create(ctx, projectID, input)
	if err != nil {
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error", err.Error())))
		span.SetStatus(codes.Error, err.Error())
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	//createdGoods, err := h.services.Goods.GetOne(id, projectID)
	//if err != nil {
	//	newErrorResponse(c, http.StatusInternalServerError, err.Error())
	//	return
	//}

	response := &models.Goods{
		ID:          id,
		ProjectID:   projectID,
		Name:        input.Name,
		Description: input.Description,
		Priority:    input.Priority,
		Removed:     input.Removed,
		CreatedAt:   time.Now(),
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

	ctx, span := h.tracer.Start(c.Request.Context(), "getAllGoods")
	defer span.End()

	goods, err := h.services.Goods.GetAll(ctx, limit, offset)
	span.AddEvent("get all goods", trace.WithAttributes(attribute.String("count", fmt.Sprint(len(goods.Goods)))))
	if err != nil {
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error", err.Error())))
		span.SetStatus(codes.Error, err.Error())
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

	ctx, span := h.tracer.Start(c.Request.Context(), "getOne")
	defer span.End()
	span.AddEvent("getOne", trace.WithAttributes(attribute.String("id", fmt.Sprint(goodsID))))

	goods, err := h.services.Goods.GetOne(ctx, goodsID, projectID)
	if err != nil {
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error", err.Error())))
		span.SetStatus(codes.Error, err.Error())
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	} else {
		p.GoodsCounter.With(prometheus.Labels{"project_id": fmt.Sprint(projectID)}).Inc()
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

	ctx, span := h.tracer.Start(c.Request.Context(), "updateGoods")
	defer span.End()
	span.AddEvent("updateGoods", trace.WithAttributes(attribute.String("goodsID", fmt.Sprintf("%d", goodsID))))

	if err := h.services.Goods.Update(ctx, goodsID, projectID, input); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			span.RecordError(err, trace.WithAttributes(
				attribute.String("error", err.Error())))
			span.SetStatus(codes.Error, err.Error())
			newDetailedErrorResponse(c, http.StatusNotFound, 3, "errors.good.NotFound", "record not found")
			return
		}
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error", err.Error())))
		span.SetStatus(codes.Error, err.Error())
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	updatedGoods, err := h.services.Goods.GetOne(ctx, goodsID, projectID)
	span.AddEvent("updated goods",
		trace.WithAttributes(
			attribute.String("id", fmt.Sprint(updatedGoods.ID)),
			attribute.String("name", updatedGoods.Name),
		))
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
	ctx, span := h.tracer.Start(c.Request.Context(), "deleteGoods")

	defer span.End()
	span.AddEvent("deleteGoods", trace.WithAttributes(attribute.String("goodsID", fmt.Sprintf("%d", goodsID))))

	if err := h.services.Goods.Delete(ctx, goodsID, projectID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			span.RecordError(err, trace.WithAttributes(
				attribute.String("error", err.Error())))
			span.SetStatus(codes.Error, err.Error())
			newDetailedErrorResponse(c, http.StatusNotFound, 3, "errors.good.NotFound", "record not found")
			return
		}
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error", err.Error())))
		span.SetStatus(codes.Error, err.Error())
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	updatedGoods, err := h.services.Goods.GetOne(ctx, goodsID, projectID)
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

	ctx, span := h.tracer.Start(c.Request.Context(), "reprioritize")
	defer span.End()
	span.AddEvent("reprioritize")

	err = h.services.Goods.Reprioritize(ctx, goodsID, projectID, priority)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			span.RecordError(err, trace.WithAttributes(
				attribute.String("error", err.Error())))
			span.SetStatus(codes.Error, err.Error())
			newDetailedErrorResponse(c, http.StatusNotFound, 3, "errors.good.NotFound", "record not found")
			return
		}
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error", err.Error())))
		span.SetStatus(codes.Error, err.Error())
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	updatedGoods, err := h.services.Goods.GetOne(ctx, goodsID, projectID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, updatedGoods)
}
