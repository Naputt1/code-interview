package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/naputt/code-interview/backend/internal/service"
)

// listOrders godoc
// @Summary  List all orders with their steps
// @Tags     orders
// @Produce  json
// @Success  200 {array} domain.Order
// @Router   /orders [get]
func (s *Server) listOrders(c *gin.Context) {
	orders, err := s.svc.ListOrders(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, orders)
}

// createOrder godoc
// @Summary  Create an order and snapshot the current process sequence into its steps
// @Tags     orders
// @Accept   json
// @Produce  json
// @Param    order body service.OrderInput true "Order payload"
// @Success  201 {object} domain.Order
// @Failure  400 {object} ErrorResponse
// @Router   /orders [post]
func (s *Server) createOrder(c *gin.Context) {
	var input service.OrderInput
	if !bindJSON(c, &input) {
		return
	}
	order, err := s.svc.CreateOrder(c.Request.Context(), input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, order)
}

// getOrder godoc
// @Summary  Get one order with steps and consumption
// @Tags     orders
// @Produce  json
// @Param    id path int true "Order ID"
// @Success  200 {object} domain.Order
// @Failure  404 {object} ErrorResponse
// @Router   /orders/{id} [get]
func (s *Server) getOrder(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	order, err := s.svc.GetOrder(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, order)
}

// cancelOrder godoc
// @Summary  Cancel an order that has no running step
// @Tags     orders
// @Produce  json
// @Param    id path int true "Order ID"
// @Success  200 {object} domain.Order
// @Failure  404 {object} ErrorResponse
// @Failure  409 {object} ErrorResponse
// @Router   /orders/{id}/cancel [post]
func (s *Server) cancelOrder(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	order, err := s.svc.CancelOrder(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, order)
}

// startStep godoc
// @Summary  Start an order's step (checks sequence, WIP limit and resource capacity)
// @Tags     orders
// @Produce  json
// @Param    id path int true "Order ID"
// @Param    stepId path int true "Order step ID"
// @Success  200 {object} domain.Order
// @Failure  404 {object} ErrorResponse
// @Failure  409 {object} ErrorResponse
// @Router   /orders/{id}/steps/{stepId}/start [post]
func (s *Server) startStep(c *gin.Context) {
	orderID, ok := pathID(c, "id")
	if !ok {
		return
	}
	stepID, ok := pathID(c, "stepId")
	if !ok {
		return
	}
	order, err := s.svc.StartStep(c.Request.Context(), orderID, stepID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, order)
}

// completeStep godoc
// @Summary  Complete an order's running step
// @Tags     orders
// @Produce  json
// @Param    id path int true "Order ID"
// @Param    stepId path int true "Order step ID"
// @Success  200 {object} domain.Order
// @Failure  404 {object} ErrorResponse
// @Failure  409 {object} ErrorResponse
// @Router   /orders/{id}/steps/{stepId}/complete [post]
func (s *Server) completeStep(c *gin.Context) {
	orderID, ok := pathID(c, "id")
	if !ok {
		return
	}
	stepID, ok := pathID(c, "stepId")
	if !ok {
		return
	}
	order, err := s.svc.CompleteStep(c.Request.Context(), orderID, stepID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, order)
}
