package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Home",
		})
	})

	// 1. Health check endpoint
	// Used to verify if the server is urnning via browser
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// 2. Github webhook receiver
	// Github will send POST requests here when a PR is opened
	r.POST("/webhook", func(c *gin.Context) {
		fmt.Println("Webhook event received!")

		c.JSON(http.StatusOK, gin.H{
			"status": "received",
		})
	})

	fmt.Println("🚀 Server is running on port 8080...")
	r.Run(":8080")
}
