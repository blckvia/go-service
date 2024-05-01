package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

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

	ctx, span := h.tracer.Start(c.Request.Context(), "createProject")
	id, err := h.services.Projects.Create(ctx, input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	defer span.End()
	span.AddEvent("created project", trace.WithAttributes(attribute.String("name", input.Name)))

	project, err := h.services.Projects.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
		span.SetStatus(codes.Error, "error")
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

	ctx, span := h.tracer.Start(c.Request.Context(), "getAllProjects")
	defer span.End()

	projects, err := h.services.Projects.GetAll(ctx, limit, offset)
	span.AddEvent("get all projects", trace.WithAttributes(attribute.String("total", fmt.Sprint(len(projects.Projects)))))
	if err != nil {
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
		span.SetStatus(codes.Error, "error")
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

	ctx, span := h.tracer.Start(c.Request.Context(), "getProject")
	defer span.End()

	span.AddEvent("get project", trace.WithAttributes(attribute.String("id", fmt.Sprint(projectID))))
	project, err := h.services.Projects.GetByID(ctx, projectID)
	if err != nil {
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
		span.SetStatus(codes.Error, "error")
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

	ctx, span := h.tracer.Start(c.Request.Context(), "updateProject")
	defer span.End()
	span.AddEvent("update project", trace.WithAttributes(attribute.String("id", fmt.Sprint(projectID))))

	if err := h.services.Projects.Update(ctx, projectID, input); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			span.RecordError(err, trace.WithAttributes(
				attribute.String("error", err.Error()),
			))
			span.SetStatus(codes.Error, "error")
			newDetailedErrorResponse(c, http.StatusNotFound, 3, "errors.project.NotFound", "record not found")
			return
		}
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
		span.SetStatus(codes.Error, "error")
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	updatedProjects, err := h.services.Projects.GetByID(ctx, projectID)
	span.AddEvent("get updated project", trace.WithAttributes(attribute.String("updated project", fmt.Sprint(updatedProjects.Name, fmt.Sprint(updatedProjects.ID)))))
	if err != nil {
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
		span.SetStatus(codes.Error, "error")
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

	ctx, span := h.tracer.Start(c.Request.Context(), "deleteProject")
	defer span.End()
	span.AddEvent("delete project", trace.WithAttributes(attribute.String("id", fmt.Sprint(projectID))))

	if err := h.services.Projects.Delete(ctx, projectID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			span.RecordError(err, trace.WithAttributes(
				attribute.String("error", err.Error()),
			))
			span.SetStatus(codes.Error, "error")
			newDetailedErrorResponse(c, http.StatusNotFound, 3, "errors.project.NotFound", "record not found")
			return
		}
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
		span.SetStatus(codes.Error, "error")
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{"ok"})
}
