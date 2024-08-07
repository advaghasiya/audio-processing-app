package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string      `gorm:"unique;not null"`
	Email        string      `gorm:"unique;not null"`
	PasswordHash string      `gorm:"not null"`
	AudioFiles   []AudioFile `gorm:"foreignKey:UserID"`
}

type AudioFile struct {
	gorm.Model
	Name          string
	Path          string
	UserID        uint
	AudioMetadata AudioMetadata `gorm:"foreignKey:AudioFileID"`
	DeviceInfo    DeviceInfo    `gorm:"foreignKey:AudioFileID"`
}

type AudioMetadata struct {
	gorm.Model
	AudioFileID         uint
	OriginalSampleRate  int
	ResampledSampleRate int
	Duration            float64
	Channels            int
	BitsPerSample       int
	ProcessingTime      float64
	IntegrityMessage    string
	DurationMessage     string
	Title               string
	Artist              string
	Album               string
	Genre               string
	Year                int
	Bitrate             int
	Filesize            int
}

type DeviceInfo struct {
	gorm.Model
	AudioFileID    uint
	Browser        string
	BrowserVersion string
	OS             string
	OSVersion      string
	Device         string
	IsMobile       bool
	IsTablet       bool
	IsPC           bool
	IPAddress      string
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}
