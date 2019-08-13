package models

import (
	"github.com/jinzhu/gorm"
	"github.com/screepsplus/screepsplus/auth"
)

// User model
type User struct {
	gorm.Model
	Username       string `gorm:"size:255;unique_index"`
	Email          string `gorm:"type:varchar(100);unique_index"`
	HashedPassword string `gorm:"type:text" json:"-"`
	Active         bool
	Verified       bool
}

// SetNewPassword set a new hashed password to user
func (user *User) SetNewPassword(passwordString string) {
	hash, _ := auth.HashPassword(passwordString)
	user.HashedPassword = hash
}

// VerifyPassword verifies a password
func (user *User) VerifyPassword(passwordString string) bool {
	valid, _ := auth.VerifyPassword(user.HashedPassword, passwordString)
	return valid
}
