package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-service/internal/models"
)

func (h *Handler) createProject(c *gin.Context) {
	var input models.Project
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.services.Projects.Create(input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})
}

type getAllProjectsResponse struct {
	Data []models.Project `json:"data"`
}

func (h *Handler) getAllProjects(c *gin.Context) {
	project, err := h.services.Projects.GetAll()
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, getAllProjectsResponse{
		Data: project,
	})
}

func (h *Handler) getProject(c *gin.Context) {
	projectID, err := GetProjectId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	project, err := h.services.Projects.GetByID(projectID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *Handler) updateProject(c *gin.Context) {
	projectID, err := GetProjectId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	var input models.Project
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.services.Projects.Update(projectID, input); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{"ok"})
}

func (h *Handler) deleteProject(c *gin.Context) {
	projectID, err := GetProjectId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	err = h.services.Projects.Delete(projectID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{"ok"})
}
