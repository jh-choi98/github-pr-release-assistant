package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const GITHUB_SIGNATURE_HEADER = "sha256="

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Fatal("WEBHOOK_SECRET not set")
	}

	r := setupRouter(webhookSecret)

	fmt.Println("Server running on :8080 with HMAC Security enabled...")
	r.Run(":8080")
}

func verifySignature(payload []byte, signatureHeader string, secret string) bool {
	if !strings.HasPrefix(signatureHeader, GITHUB_SIGNATURE_HEADER) {
		return false
	}

	signature := strings.TrimPrefix(signatureHeader, GITHUB_SIGNATURE_HEADER)
	
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

func setupRouter(webhookSecret string) *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	r.POST("/webhook", func(c *gin.Context) {
		payloadBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		signature := c.GetHeader("X-Hub-Signature-256")
		if !verifySignature(payloadBody, signature, webhookSecret) {
			fmt.Println("Security Alert: Signature verification failed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
			return
		}

		var event GitHubPullRequestEvent
		if err := json.Unmarshal(payloadBody, &event); err != nil {
			fmt.Println("JSON Parsing error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		if event.Action == "opened" {
			fmt.Printf("\n🚀 New PR Detected!\n")
			fmt.Printf("	Repo:	%s\n", event.Repository.FullName)
			fmt.Printf("	PR #%d:	%s\n", event.Number, event.PullRequest.Title)

			go HandleNewPR(event)
		} else {
			fmt.Printf("Ignored event type: %s\n", event.Action)
		}

		c.JSON(http.StatusOK, gin.H{"status": "processed"})
	})

	return r
}
