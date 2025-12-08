package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint      `gorm:"primaryKey"`
	Username     string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time
}

type Repository struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"not null"`
	Path      string    `gorm:"uniqueIndex;not null"`
	UserID    uint      `gorm:"not null"`
	CreatedAt time.Time
}

// Migrate runs the auto-migration for the models
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Repository{})
}
