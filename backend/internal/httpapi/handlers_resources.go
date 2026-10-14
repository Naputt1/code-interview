package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/naputt/code-interview/backend/internal/service"
)

// listResources godoc
// @Summary  List all resources
// @Tags     resources
// @Produce  json
// @Success  200 {array} domain.Resource
// @Router   /resources [get]
func (s *Server) listResources(c *gin.Context) {
	resources, err := s.svc.ListResources(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resources)
}

// createResource godoc
// @Summary  Create a resource
// @Tags     resources
// @Accept   json
// @Produce  json
// @Param    resource body service.ResourceInput true "Resource payload"
// @Success  201 {object} domain.Resource
// @Failure  400 {object} ErrorResponse
// @Failure  409 {object} ErrorResponse
// @Router   /resources [post]
func (s *Server) createResource(c *gin.Context) {
	var input service.ResourceInput
	if !bindJSON(c, &input) {
		return
	}
	resource, err := s.svc.CreateResource(c.Request.Context(), input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resource)
}

// updateResource godoc
// @Summary  Update a resource
// @Tags     resources
// @Accept   json
// @Produce  json
// @Param    id path int true "Resource ID"
// @Param    resource body service.ResourceInput true "Resource payload"
// @Success  200 {object} domain.Resource
// @Failure  400 {object} ErrorResponse
// @Failure  404 {object} ErrorResponse
// @Failure  409 {object} ErrorResponse
// @Router   /resources/{id} [put]
func (s *Server) updateResource(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var input service.ResourceInput
	if !bindJSON(c, &input) {
		return
	}
	resource, err := s.svc.UpdateResource(c.Request.Context(), id, input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resource)
}

// deleteResource godoc
// @Summary  Delete a resource
// @Tags     resources
// @Produce  json
// @Param    id path int true "Resource ID"
// @Success  204
// @Failure  404 {object} ErrorResponse
// @Failure  409 {object} ErrorResponse
// @Router   /resources/{id} [delete]
func (s *Server) deleteResource(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := s.svc.DeleteResource(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
