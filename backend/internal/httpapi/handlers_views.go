package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// board godoc
// @Summary  Orders currently in progress at each step
// @Tags     views
// @Produce  json
// @Success  200 {object} service.Board
// @Router   /board [get]
func (s *Server) board(c *gin.Context) {
	board, err := s.svc.Board(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, board)
}

// overview godoc
// @Summary  Status of all orders, per-step load and resource consumption
// @Tags     views
// @Produce  json
// @Success  200 {object} service.Overview
// @Router   /overview [get]
func (s *Server) overview(c *gin.Context) {
	overview, err := s.svc.Overview(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, overview)
}

// listConsumptions godoc
// @Summary  Recent entries from the consumption ledger
// @Tags     views
// @Produce  json
// @Success  200 {array} domain.Consumption
// @Router   /consumptions [get]
func (s *Server) listConsumptions(c *gin.Context) {
	entries, err := s.svc.ListConsumptions(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, entries)
}
