package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/advaghasiya/audio-processing-app/internal/config"
	"github.com/advaghasiya/audio-processing-app/internal/handlers"
	"github.com/advaghasiya/audio-processing-app/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfg := config.New()

	// Ensure upload folder exists
	if err := os.MkdirAll(cfg.UploadFolder, os.ModePerm); err != nil {
		log.Fatalf("Failed to create upload folder: %v", err)
	}

	db, err := gorm.Open(sqlite.Open(cfg.SQLiteDatabaseURI), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.AudioFile{}, &models.AudioMetadata{}, &models.DeviceInfo{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	r := gin.Default()

	// Correct path construction for template files
	templatesPath := filepath.Join(cfg.BaseDir, "web", "templates", "*")
	r.LoadHTMLGlob(templatesPath)

	r.Static("/static", filepath.Join(cfg.BaseDir, "web", "static"))

	h := handlers.New(db, cfg)

	r.GET("/", h.Index)
	r.GET("/login", h.LoginPage)
	r.POST("/api/login", h.Login)
	r.GET("/signup", h.SignupPage)
	r.POST("/api/signup", h.Signup)
	r.GET("/:username", h.Dashboard)

	authorized := r.Group("/")
	authorized.Use(authMiddleware(cfg.JWTSecretKey))
	{
		authorized.POST("/api/upload", h.UploadFile)
		authorized.GET("/api/audio-files", h.GetAudioFiles)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func authMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse token claims"})
			c.Abort()
			return
		}

		c.Set("user_id", claims["user_id"])
		c.Next()
	}
}
