package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	SecretKey          string
	BaseDir            string
	UploadFolder       string
	AllowedExtensions  map[string]bool
	MaxContentLength   int64
	AllowedSampleRates map[int]bool
	TargetSampleRate   int
	SQLiteDatabaseURI  string
	JWTSecretKey       string
}

func New() *Config {
	baseDir, _ := os.Getwd()
	return &Config{
		SecretKey:          getEnv("SECRET_KEY", "hard-to-guess-string"),
		BaseDir:            baseDir,
		UploadFolder:       filepath.Join(baseDir, "uploads"),
		AllowedExtensions:  map[string]bool{"wav": true},
		MaxContentLength:   32 * 1024 * 1024, // 32 MB limit
		AllowedSampleRates: map[int]bool{1000: true, 2000: true, 4000: true, 8000: true, 16000: true, 22050: true, 44100: true, 48000: true},
		TargetSampleRate:   8000,
		SQLiteDatabaseURI:  filepath.Join(baseDir, "app.db"),
		JWTSecretKey:       getEnv("JWT_SECRET_KEY", "jwt-secret-string"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
