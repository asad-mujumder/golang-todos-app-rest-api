package router

import (
	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	router := gin.Default()
	router.SetTrustedProxies(nil)

	// Health Cheack - Base Path
	router.GET("/", func(ctx *gin.Context){
		ctx.JSON(200, gin.H{
			"success": true,
			"Message": "API is working",
		})
	})

	return router
}