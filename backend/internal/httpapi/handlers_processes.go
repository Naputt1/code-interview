package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/naputt/code-interview/backend/internal/service"
)

// listProcesses godoc
// @Summary  List processes in sequence order
// @Tags     processes
// @Produce  json
// @Success  200 {array} domain.Process
// @Router   /processes [get]
func (s *Server) listProcesses(c *gin.Context) {
	processes, err := s.svc.ListProcesses(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, processes)
}

// createProcess godoc
// @Summary  Create a process (step) with its resource recipe
// @Tags     processes
// @Accept   json
// @Produce  json
// @Param    process body service.ProcessInput true "Process payload"
// @Success  201 {object} domain.Process
// @Failure  400 {object} ErrorResponse
// @Failure  409 {object} ErrorResponse
// @Router   /processes [post]
func (s *Server) createProcess(c *gin.Context) {
	var input service.ProcessInput
	if !bindJSON(c, &input) {
		return
	}
	process, err := s.svc.CreateProcess(c.Request.Context(), input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, process)
}

// getProcess godoc
// @Summary  Get one process
// @Tags     processes
// @Produce  json
// @Param    id path int true "Process ID"
// @Success  200 {object} domain.Process
// @Failure  404 {object} ErrorResponse
// @Router   /processes/{id} [get]
func (s *Server) getProcess(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	process, err := s.svc.GetProcess(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, process)
}

// updateProcess godoc
// @Summary  Update a process, its recipe and/or its place in the sequence
// @Tags     processes
// @Accept   json
// @Produce  json
// @Param    id path int true "Process ID"
// @Param    process body service.ProcessInput true "Process payload"
// @Success  200 {object} domain.Process
// @Failure  400 {object} ErrorResponse
// @Failure  404 {object} ErrorResponse
// @Failure  409 {object} ErrorResponse
// @Router   /processes/{id} [put]
func (s *Server) updateProcess(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var input service.ProcessInput
	if !bindJSON(c, &input) {
		return
	}
	process, err := s.svc.UpdateProcess(c.Request.Context(), id, input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, process)
}

// deleteProcess godoc
// @Summary  Delete a process
// @Tags     processes
// @Produce  json
// @Param    id path int true "Process ID"
// @Success  204
// @Failure  404 {object} ErrorResponse
// @Failure  409 {object} ErrorResponse
// @Router   /processes/{id} [delete]
func (s *Server) deleteProcess(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := s.svc.DeleteProcess(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
