package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-service/internal/models"
	"go-service/internal/repository"
)

// @Summary Create project
// @Tags Projects
// @Description Create project
// @ID create-project
// @Accept  json
// @Produce  json
// @Param input body models.UpdateProject true "project info"
// @Success 200 {object} models.Goods
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/projects [post]
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

// @Summary Get all projects
// @Tags Projects
// @Description Get all projects
// @ID get-all-projects
// @Accept  json
// @Produce  json
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {object} models.GetAllProjects
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/projects [get]
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

// @Summary Get project
// @Tags Projects
// @Description Get project
// @ID get-project
// @Accept  json
// @Produce  json
// @Param id path int true "project_id"
// @Success 200 {object} models.Project
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/projects/{id} [get]
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

// @Summary Update project
// @Tags Projects
// @Description Get project By ID
// @ID update-project
// @Accept  json
// @Produce  json
// @Param input body models.UpdateProject true "project info"
// @Param id path int true "project_id"
// @Success 200 {object} models.Project
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/projects/{id} [patch]
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

// @Summary Delete project
// @Tags Projects
// @Description Delete project
// @ID delete-project
// @Accept  json
// @Produce  json
// @Param id path int true "project_id"
// @Success 200
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/projects/{id} [delete]
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
