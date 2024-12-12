package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	HEALTH_DURATION_SEC = 60
)

type NewConnectionInput struct {
	URL string `json:"url" binding:"required"`
}

type RemoveConnectionInput struct {
	id string `json:"id" binding:"required"`
}

func allConnections(c *gin.Context, lb *LoadBalancer) {
	c.JSON(http.StatusOK, gin.H{
		"connnections": lb.connections,
	})
}

func addConnection(c *gin.Context, lb *LoadBalancer) {
	var input NewConnectionInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lb.addConnection(input.URL)
	c.JSON(http.StatusOK, gin.H{"message": "Success", "url": input.URL})
}

func removeConnection(c *gin.Context, lb *LoadBalancer) {
	var input RemoveConnectionInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lb.removeConnection(input.id)
	c.JSON(http.StatusOK, gin.H{"message": "Success", "url": input.id})
}

func main() {

	lb := newLoadBalancer()

	go lb.initHealthCheck(HEALTH_DURATION_SEC)
	router := gin.Default()

	router.GET("/", func(ctx *gin.Context) {
		allConnections(ctx, lb)
	})

	router.POST("/lb/add-connection", func(ctx *gin.Context) {
		addConnection(ctx, lb)
	})

	router.POST("/lb/remove-connection", func(ctx *gin.Context) {
		removeConnection(ctx, lb)
	})

	router.Run(":8080")
}
