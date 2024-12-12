package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	HEALTH_DURATION_SEC = 10
)

type NewConnectionInput struct {
	URL string `json:"url" binding:"required"`
}

type RemoveConnectionInput struct {
	ID string `json:"id" binding:"required"`
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

	lb.removeConnection(input.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Success", "url": input.ID})
}

func makeRequestToBackend(c *gin.Context, lb *LoadBalancer) {

	method := c.Request.Method
	path := c.Request.URL.Path
	query := c.Request.URL.RawQuery
	headers := c.Request.Header

	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
	}

	connection, err := lb.getNextTarget()

	if err != nil {
		fmt.Print(err)
		c.JSON(500, gin.H{"error": "Error finding target"})
		return
	}

	backendURL := connection.URL + path

	if query != "" {
		backendURL += "?" + query
	}

	fmt.Printf("Path %s \n", path)
	fmt.Printf("Formatted Backend Url %s \n", backendURL)

	req, err := http.NewRequest(method, backendURL, bytes.NewReader(bodyBytes))
	if err != nil {
		fmt.Print(err)
		c.JSON(500, gin.H{"error": "Failed to create request"})
		return
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to forward request to backend"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)

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

	router.NoRoute(func(c *gin.Context) {
		makeRequestToBackend(c, lb)
	})

	router.Run(":8080")
}
