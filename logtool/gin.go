package logtool

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Gin middleware for zap logging
func GinZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		errs := c.Errors.ByType(gin.ErrorTypePrivate).String()
		title := http.StatusText(status)

		addBody := false
		var logFn func(string, ...interface{})
		switch {
		case status >= 500:
			logFn = GetLogger().Errorw
			addBody = true
		case status >= 400:
			logFn = GetLogger().Warnw
			addBody = true
		default:
			logFn = GetLogger().Infow
		}

		fields := []interface{}{
			"method", method,
			"path", path,
			"status", status,
			"caller", "logtool.GinZapLogger",
		}

		if !DevMode {
			fields = append(fields,
				"latency", latency.String(),
				"ip", clientIP,
				"user_agent", userAgent,
			)
		}

		if errs != "" {
			fields = append(fields, "error", errs)
		}

		// Optional: Read response body only if explicitly captured earlier (since Gin doesn't expose it directly)
		if addBody {
			// Requires a response body writer wrapper if you really need to log the response body
			fields = append(fields, "body", "(response body logging not implemented in Gin by default)")
		}

		logFn(title, fields...)
	}
}
