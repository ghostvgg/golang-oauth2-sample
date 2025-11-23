package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
)

func main() {

	// -------------------------
	// OAuth2 Manager & Stores
	// -------------------------
	manager := manage.NewDefaultManager()

	// Token configuration (ADD refresh tokens)
	manager.SetAuthorizeCodeTokenCfg(&manage.Config{
		AccessTokenExp:    time.Hour * 2,       // access token expiration
		RefreshTokenExp:   time.Hour * 24 * 30, // refresh token expiration (30 days)
		IsGenerateRefresh: true,                // <-- IMPORTANT
	})

	// Token store (in-memory for demo)
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	// Client store (in-memory for demo)
	clientStore := store.NewClientStore()
	clientStore.Set("demo-client", &models.Client{
		ID:     "demo-client",
		Secret: "demo-secret",
		Domain: "http://localhost:8080/callback",
	})
	manager.MapClientStorage(clientStore)

	// -------------------------
	// OAuth2 Server
	// -------------------------
	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)

	// Receive client_id + secret via form
	srv.SetClientInfoHandler(server.ClientFormHandler)

	// Optional: Add logs for debugging
	srv.SetInternalErrorHandler(func(err error) *errors.Response {
		log.Println("Internal Error:", err.Error())
		return nil
	})
	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	srv.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (userID string, err error) {
		// Simulate logged-in user
		return "demo-user", nil
	})

	r := gin.Default()

	// -------------------------
	// /authorize (returns code)
	// -------------------------
	r.GET("/authorize", func(c *gin.Context) {
		err := srv.HandleAuthorizeRequest(c.Writer, c.Request)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
		}
	})

	// -------------------------
	// /token (access + refresh)
	// -------------------------
	r.POST("/token", func(c *gin.Context) {
		srv.HandleTokenRequest(c.Writer, c.Request)
	})

	r.GET("/callback", func(c *gin.Context) {
		code := c.Query("code")
		c.JSON(200, gin.H{
			"message": "Received authorization code",
			"code":    code,
		})
	})

	// -------------------------
	// Protected API
	// -------------------------
	r.GET("/hello", func(c *gin.Context) {
		ti, err := srv.ValidationBearerToken(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"message": "Hello from protected API",
			"user":    ti.GetUserID(),
			"scope":   ti.GetScope(),
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Println("OAuth2 demo server with refresh token running on :8080")
	r.Run(":8080")
}
