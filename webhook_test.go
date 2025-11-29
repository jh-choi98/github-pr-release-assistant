package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func generateTestSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// ---------------------------------------------------------
// Part 1: Unit Test for Security Logic (Table-Driven)
// ---------------------------------------------------------
func TestVerifySignature(t *testing.T) {
	secret := "test-secret"
	payload := []byte(`{"action": "opened"}`)

	tests := []struct {
		name			string
		payload			[]byte
		signature		string
		secret			string
		expectedResult	bool
	}{
		{
			name:			"✅ Valid Signature",
			payload:		payload,
			signature:		generateTestSignature(payload, secret),
			secret:			secret,
			expectedResult: true,
		},
		{
			name:			"❌ Invalid Secret (Hacker)",
			payload:		payload,
			signature:		generateTestSignature(payload, "wrong-secret"),
			secret:			secret,
			expectedResult: false,
		},
		{
			name:			"❌ Tampered Payload (Man-in-the-Middle)",
			payload:		[]byte(`{"action":"closed"}`),
			signature:		generateTestSignature(payload, secret),
			secret:			secret,
			expectedResult: false,
		},
		{
			name:			"❌ Malformed Header",
			payload:		payload,
			signature:		"invalid-format",
			secret:			secret,
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := verifySignature(test.payload, test.signature, test.secret)
			assert.Equal(t, test.expectedResult, result)
		})
	}
}

// ---------------------------------------------------------
// Part 2: Integration Test for the HTTP Handler
// This tests Parsing + Security + Response Codes together
// ---------------------------------------------------------
func TestWebhookHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testSecret := "test-secret-for-unit-test"

	router := setupRouter(testSecret)

	tests := []struct {
		name			string
		payload			string
		setupHeaders	func(req *http.Request, body []byte)
		expectedStatus	int
	}{
		{
			name: "✅ Happy Path (Authorized)",
			payload: `{"action": "opened", "number": 1}`,
			setupHeaders: func(req *http.Request, body []byte) {
				sig := generateTestSignature(body, testSecret)
				req.Header.Set("X-Hub-Signature-256", sig)
				req.Header.Set("Content-Type", "application/json")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "❌ Unauthorized (Bad Signature)",
			payload: `{"action": "opened"}`,
			setupHeaders: func(req *http.Request, body []byte) {
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Hub-Signature-256", "sha256=badsignature")
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := []byte(test.payload)
			req, _ := http.NewRequest("POST", "/webhook", bytes.NewBuffer(body))
			test.setupHeaders(req, body)

			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
		})
	}
}
