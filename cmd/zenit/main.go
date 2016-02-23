package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"bytes"
	"io/ioutil"
	"crypto/sha1"
	"crypto/hmac"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	"github.com/octokit/go-octokit/octokit"
)

type RequestBodyReader struct {
	*bytes.Buffer
}
func (reader RequestBodyReader) Close() error { return nil }

type Pusher struct {
	Name string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

type PushEvent struct {
	Pusher Pusher `json:"pusher" binding:"required"`
}

// Reads the request body in a buffer and replaces context.Request.Body with a
// new buffer so that it can be read again by subsequent consumers.
func GetRequestBody(context *gin.Context) []byte {
	buffer, err := ioutil.ReadAll(context.Request.Body)
	if err != nil {
		log.Fatal(err)
	}

	newReader := RequestBodyReader{bytes.NewBuffer(buffer)}
	context.Request.Body = newReader

	return buffer
}

func CheckSecret(context *gin.Context) bool {
	bodyContent := GetRequestBody(context)

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
	if CheckSecret(context) {
		var push PushEvent
		context.Bind(&push)

		response := fmt.Sprintf("Handling a push event:\n\n%+v", push)
		context.String(http.StatusOK, response)
	}
}

func HandlePullRequest(context *gin.Context, client *octokit.Client) {
	if CheckSecret(context) {
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
