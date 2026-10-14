package httpapi

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/naputt/code-interview/backend/internal/service"
)

type Server struct {
	svc *service.Service
	log *slog.Logger
}

// New builds the gin engine and registers every route.
func New(svc *service.Service, log *slog.Logger) *gin.Engine {
	s := &Server{svc: svc, log: log}

	router := gin.New()
	router.Use(gin.Recovery(), requestLogger(log), corsMiddleware())

	api := router.Group("/api")
	{
		api.GET("/health", s.health)

		api.GET("/resources", s.listResources)
		api.POST("/resources", s.createResource)
		api.PUT("/resources/:id", s.updateResource)
		api.DELETE("/resources/:id", s.deleteResource)

		api.GET("/processes", s.listProcesses)
		api.POST("/processes", s.createProcess)
		api.GET("/processes/:id", s.getProcess)
		api.PUT("/processes/:id", s.updateProcess)
		api.DELETE("/processes/:id", s.deleteProcess)

		api.GET("/orders", s.listOrders)
		api.POST("/orders", s.createOrder)
		api.GET("/orders/:id", s.getOrder)
		api.POST("/orders/:id/cancel", s.cancelOrder)
		api.POST("/orders/:id/steps/:stepId/start", s.startStep)
		api.POST("/orders/:id/steps/:stepId/complete", s.completeStep)

		api.GET("/board", s.board)
		api.GET("/overview", s.overview)
		api.GET("/consumptions", s.listConsumptions)
	}
	return router
}

// health godoc
// @Summary  Liveness probe
// @Tags     system
// @Produce  json
// @Success  200 {object} map[string]string
// @Router   /health [get]
func (s *Server) health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
