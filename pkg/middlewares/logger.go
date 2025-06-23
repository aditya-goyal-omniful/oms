package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omniful/go_commons/log"
)

func RequestLogger() gin.HandlerFunc {
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

		log.Infof("Method=%s Path=%s Query=%s Status=%d Latency=%s TenantID=%s",
			method, path, raw, statusCode, latency, tenantID)
	}
}
