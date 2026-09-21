package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 打印每个请求的方法、路径、状态码与耗时。
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("[http] %s %s -> %d (%s)",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start))
	}
}
