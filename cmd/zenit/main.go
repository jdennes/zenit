package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func handlePush(c *gin.Context) {
	c.String(http.StatusOK, "Handling a push event")
}

func handlePullRequest(c *gin.Context) {
	c.String(http.StatusOK, "Handling a pull_request event")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	router.POST("/handle", func(c *gin.Context) {
		event := c.Request.Header["X-Github-Event"][0]
		switch event {
		case "push":
			handlePush(c)
		case "pull_request":
			handlePullRequest(c)
		default:
			log.Fatal("Unsupported event in the X-Github-Event HTTP header")
		}
	})

	router.Run(":" + port)
}
