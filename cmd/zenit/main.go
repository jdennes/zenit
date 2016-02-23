package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"io/ioutil"
	"crypto/sha1"
	"crypto/hmac"
	"encoding/hex"

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

func CheckSecret(context *gin.Context, bodyContent []byte) bool {
	// X-Hub-Signature header format:
	// https://developer.github.com/webhooks/securing/#validating-payloads-from-github
	header := context.Request.Header.Get("X-Hub-Signature")
	mac := hmac.New(sha1.New, []byte(os.Getenv("SECRET")))
	mac.Write(bodyContent)
	expectedMAC := mac.Sum(nil)
	signature := fmt.Sprintf("sha1=%s", hex.EncodeToString(expectedMAC))

	if !hmac.Equal([]byte(header), []byte(signature)) {
		context.String(http.StatusForbidden, "Unacceptable X-Hub-Signature HTTP header")
		return false
	}
	return true
}

func HandlePush(context *gin.Context, client *octokit.Client) {
	var _ Push

	// TODO: Figure out how to let CheckSecret read the raw value of
	// context.Request.Body and have the binding still work.
	// context.BindJSON(&push)

	bodyContent, _ := ioutil.ReadAll(context.Request.Body)
	if CheckSecret(context, bodyContent) {
		context.String(http.StatusOK, "Handling a push event")
	}
}

func HandlePullRequest(context *gin.Context, client *octokit.Client) {
	bodyContent, _ := ioutil.ReadAll(context.Request.Body)
	if CheckSecret(context, bodyContent) {
		context.String(http.StatusOK, "Handling a pull_request event")
	}
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
			HandlePush(context, client)
		case "pull_request":
			HandlePullRequest(context, client)
		default:
			log.Fatal("Unsupported event in the X-Github-Event HTTP header")
		}
	})

	router.Run(":" + port)
}
