package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"crypto/sha256"
	"crypto/hmac"

	"github.com/gin-gonic/gin"
	"github.com/octokit/go-octokit/octokit"
)

type Pusher struct {
	Name string `form:"name" json:"name" binding:"required"`
  Email string `form:"email" json:"email" binding:"required"`
}

type Push struct {
	Pusher Pusher `form:"pusher" json:"pusher" binding:"required"`
}

func checkSecret(context *gin.Context) bool {
	// X-Hub-Signature header format:
	// https://github.com/github/github-services/blob/8dc2328d0d97005e6431c7ca8c7de9466e38567e/lib/service/http_helper.rb#L76-L77
	header := context.Request.Header.Get("X-Hub-Signature")
	mac := hmac.New(sha256.New, []byte(os.Getenv("SECRET")))
	body := ""
	mac.Write([]byte(body))
	expectedMAC := mac.Sum(nil)

	if !hmac.Equal([]byte(header), []byte(fmt.Sprintf("sha1=%v", expectedMAC))) {
		context.String(http.StatusForbidden, "Unacceptable X-Hub-Signature HTTP header")
		return false
	}
	return true
}

func handlePush(context *gin.Context, client *octokit.Client) {
	var push Push
	context.BindJSON(&push)

	if !checkSecret(context) {
		return
	}

	context.String(http.StatusOK, "Handling a push event")
}

func handlePullRequest(context *gin.Context, client *octokit.Client) {
	if !checkSecret(context) {
		return
	}
	context.String(http.StatusOK, "Handling a pull_request event")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	client := octokit.NewClient(octokit.TokenAuth{AccessToken: os.Getenv("TOKEN")})

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")

	router.GET("/", func(context *gin.Context) {
		context.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	router.POST("/handle", func(context *gin.Context) {
		event := context.Request.Header.Get("X-Github-Event")
		switch event {
		case "push":
			handlePush(context, client)
		case "pull_request":
			handlePullRequest(context, client)
		default:
			log.Fatal("Unsupported event in the X-Github-Event HTTP header")
		}
	})

	router.Run(":" + port)
}
