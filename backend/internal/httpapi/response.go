package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/naputt/code-interview/backend/internal/domain"
)

// ErrorDetail is the machine-readable error payload.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse is the envelope returned for every non-2xx response.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

func writeError(c *gin.Context, err error) {
	var derr *domain.Error
	if errors.As(err, &derr) {
		c.JSON(derr.Status, ErrorResponse{Error: ErrorDetail{Code: derr.Code, Message: derr.Message}})
		return
	}
	slog.Error("unhandled error", "error", err, "path", c.Request.URL.Path)
	c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrorDetail{
		Code:    "internal_error",
		Message: "internal server error",
	}})
}

func bindJSON(c *gin.Context, target any) bool {
	decoder := json.NewDecoder(c.Request.Body)
	// Reject unknown fields so a misspelled key fails loudly instead of being
	// silently dropped (a typo in "capacity" would otherwise mean "unlimited").
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeError(c, domain.Validation("invalid request body: "+err.Error()))
		return false
	}
	// Same struct validation Gin's default binder would run, so the
	// `binding:"required"` tags keep their meaning (and the OpenAPI contract).
	if binding.Validator != nil {
		if err := binding.Validator.ValidateStruct(target); err != nil {
			writeError(c, domain.Validation("invalid request body: "+err.Error()))
			return false
		}
	}
	return true
}

func pathID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, domain.Validation("invalid "+name))
		return 0, false
	}
	return id, true
}

func requestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info("http",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", time.Since(start).String(),
		)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if origin := c.GetHeader("Origin"); origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
