package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-service/internal/models"
)

func (h *Handler) createGoods(c *gin.Context) {
	projectID, err := GetProjectId(c)

	var input models.Goods
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.services.Goods.Create(projectID, input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})
}

type getAllGoodsResponse struct {
	Data []models.Goods `json:"data"`
}

func (h *Handler) getAllGoods(c *gin.Context) {
	goods, err := h.services.Goods.GetAll()
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, getAllGoodsResponse{
		Data: goods,
	})
}

func (h *Handler) getGoodsById(c *gin.Context) {
	goodsID, err := GetGoodsId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	goods, err := h.services.Goods.GetByID(goodsID)
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

	var input models.Goods
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.services.Goods.Update(goodsID, input); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{"ok"})
}

func (h *Handler) deleteGoods(c *gin.Context) {
	goodsID, err := GetGoodsId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	err = h.services.Goods.Delete(goodsID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{"ok"})
}
