package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	
	// 健康检查端点
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	r.Run(":8080")
}