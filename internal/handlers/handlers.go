package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/advaghasiya/audio-processing-app/internal/audio"
	"github.com/advaghasiya/audio-processing-app/internal/config"
	"github.com/advaghasiya/audio-processing-app/internal/device"
	"github.com/advaghasiya/audio-processing-app/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type Handler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func New(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{DB: db, Cfg: cfg}
}

func (h *Handler) Index(c *gin.Context) {
	c.Redirect(http.StatusFound, "/login")
}

func (h *Handler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func (h *Handler) Login(c *gin.Context) {
	var loginData struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", loginData.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.CheckPassword(loginData.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.Cfg.JWTSecretKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString, "username": user.Username})
}

func (h *Handler) SignupPage(c *gin.Context) {
	c.HTML(http.StatusOK, "signup.html", gin.H{})
}

func (h *Handler) Signup(c *gin.Context) {
	var signupData struct {
		Username        string `json:"username" binding:"required"`
		Email           string `json:"email" binding:"required,email"`
		Password        string `json:"password" binding:"required,min=6"`
		ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
	}

	if err := c.ShouldBindJSON(&signupData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	if err := h.DB.Where("email = ?", signupData.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
		return
	}

	newUser := models.User{
		Username: signupData.Username,
		Email:    signupData.Email,
	}
	if err := newUser.SetPassword(signupData.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not set password"})
		return
	}

	if err := h.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Signup successful"})
}

func (h *Handler) Dashboard(c *gin.Context) {
	username := c.Param("username")
	var user models.User
	if err := h.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	c.HTML(http.StatusOK, "dashboard.html", gin.H{"username": username})
}

func (h *Handler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	if file.Size > h.Cfg.MaxContentLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds the limit"})
		return
	}

	ext := filepath.Ext(file.Filename)
	if !h.Cfg.AllowedExtensions[ext[1:]] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File type not allowed"})
		return
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(h.Cfg.UploadFolder, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save file"})
		return
	}

	userID, _ := c.Get("user_id")

	audioInfo, err := audio.ProcessAudio(filePath, h.Cfg.TargetSampleRate, h.Cfg.AllowedSampleRates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not process audio"})
		return
	}

	deviceInfo := device.GetDeviceInfo(c.Request)

	audioFile := models.AudioFile{
		Name:   filename,
		Path:   filePath,
		UserID: userID.(uint),
		AudioMetadata: models.AudioMetadata{
			OriginalSampleRate:  audioInfo.OriginalSampleRate,
			ResampledSampleRate: audioInfo.ResampledSampleRate,
			Duration:            audioInfo.Duration,
			Channels:            audioInfo.Channels,
			BitsPerSample:       audioInfo.BitsPerSample,
			ProcessingTime:      audioInfo.ProcessingTime,
			IntegrityMessage:    audioInfo.IntegrityMessage,
			DurationMessage:     audioInfo.DurationMessage,
			Title:               audioInfo.Title,
			Artist:              audioInfo.Artist,
			Album:               audioInfo.Album,
			Genre:               audioInfo.Genre,
			Year:                audioInfo.Year,
			Bitrate:             audioInfo.Bitrate,
			Filesize:            int(file.Size),
		},
		DeviceInfo: models.DeviceInfo{
			Browser:        deviceInfo.Browser,
			BrowserVersion: deviceInfo.BrowserVersion,
			OS:             deviceInfo.OS,
			OSVersion:      deviceInfo.OSVersion,
			Device:         deviceInfo.Device,
			IsMobile:       deviceInfo.IsMobile,
			IsTablet:       deviceInfo.IsTablet,
			IsPC:           deviceInfo.IsPC,
			IPAddress:      deviceInfo.IPAddress,
		},
	}

	if err := h.DB.Create(&audioFile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save file info to database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded and processed successfully"})
}

func (h *Handler) GetAudioFiles(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var audioFiles []models.AudioFile
	if err := h.DB.Where("user_id = ?", userID).Find(&audioFiles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch audio files"})
		return
	}

	var response []gin.H
	for _, file := range audioFiles {
		response = append(response, gin.H{
			"id":   file.ID,
			"name": file.Name,
		})
	}

	c.JSON(http.StatusOK, response)
}
