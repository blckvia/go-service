package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-service/internal/models"
	"go-service/internal/repository"
)

func (h *Handler) createProject(c *gin.Context) {
	var input models.Project
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if input.Name == "" {
		newErrorResponse(c, http.StatusBadRequest, "name is required")
		return
	}

	id, err := h.services.Projects.Create(input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	project, err := h.services.Projects.GetByID(id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, project)
}

type getAllProjectsResponse struct {
	Data []models.Project `json:"data"`
}

func (h *Handler) getAllProjects(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	projects, err := h.services.Projects.GetAll(limit, offset)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, projects)
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

	var input models.UpdateProjects
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.services.Projects.Update(projectID, input); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			newDetailedErrorResponse(c, http.StatusNotFound, 3, "errors.project.NotFound", "record not found")
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	updatedProjects, err := h.services.Projects.GetByID(projectID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, updatedProjects)
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
