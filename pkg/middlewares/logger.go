package middlewares

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omniful/go_commons/i18n"
	"github.com/omniful/go_commons/log"
)

// RequestLogger logs details about each incoming HTTP request including:
// - HTTP method
// - Request path and query parameters
// - HTTP response status code
// - Time taken to serve the request
// - X-Tenant-ID header (for multi-tenancy tracking)
//
// This middleware should be registered globally to capture all requests.
//
// Example log output:
// Method=GET Path=/orders Query=status=new_order Status=200 Latency=42.34ms TenantID=abc-123
//
// Usage:
//   router.Use(middlewares.RequestLogger())
func RequestLogger(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		method := c.Request.Method
		tenantID := c.GetHeader("X-Tenant-ID")

		// Proceed to next handler
		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		log.Infof(i18n.Translate(ctx, "Method=%s Path=%s Query=%s Status=%d Latency=%s TenantID=%s"),
			method, path, raw, statusCode, latency, tenantID)
	}
}
