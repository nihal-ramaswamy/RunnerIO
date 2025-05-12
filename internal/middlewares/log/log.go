package log_middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DefaultStructuredLogger logs a gin HTTP request in JSON format. Uses the
// default logger from rs/zerolog.
func DefaultStructuredLogger(log *zap.Logger) gin.HandlerFunc {
	return StructuredLogger(log)
}

// StructuredLogger logs a gin HTTP request in JSON format. Allows to set the
// logger for testing purposes.
func StructuredLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		start := time.Now() // Start timer
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Fill the params
		param := gin.LogFormatterParams{}

		param.TimeStamp = time.Now() // Stop timer
		param.Latency = param.TimeStamp.Sub(start)

		param.ClientIP = c.ClientIP()
		param.Method = c.Request.Method
		param.StatusCode = c.Writer.Status()
		param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
		if raw != "" {
			path = path + "?" + raw
		}
		param.Path = path

		logger.Info(
			"Request Details",
			zap.String("client_id", param.ClientIP),
			zap.String("method", param.Method),
			zap.Int("status_code", param.StatusCode),
			zap.String("path", param.Path),
			zap.Duration("latency", param.Latency),
			zap.String("error_message", param.ErrorMessage),
			zap.String("user_agent", c.Request.UserAgent()),
		)
	}
}
