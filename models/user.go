package models

import (
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/screepsplus/screepsplus/auth"
)

// User model
type User struct {
	gorm.Model
	Username        string  `gorm:"size:255;unique_index"`
	Email           *string `gorm:"type:varchar(100);unique_index"`
	HashedPassword  string  `gorm:"type:text" json:"-"`
	Active          bool
	Verified        bool
	RecoverSelector string `gorm:"type:varchar(100);index"`
	RecoverVerifier string `gorm:"type:varchar(100)"`
	RecoverExpiry   time.Time
}

// NewUser Creates a new, empty user
func NewUser() *User {
	return &User{
		Email:    nil,
		Verified: false,
		Active:   true,
	}
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

// GetPID Gets the primary ID for the user (Authboss needs this)
func (user User) GetPID() string {
	return user.Username
}

// PutPID Puts the primary ID for the user (Authboss needs this)
func (user *User) PutPID(pid string) {
	user.Username = strings.ToLower(pid)
}

// GetPassword Gets the hashed password for the user
func (user User) GetPassword() string {
	return user.HashedPassword
}

// PutPassword Puts the hashed password for the user
func (user *User) PutPassword(pass string) {
	user.HashedPassword = pass
}

// GetArbitrary Gets arbitrary data
func (user User) GetArbitrary() map[string]string {
	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	return map[string]string{
		"email": email,
	}
}

// PutArbitrary Puts arbitrary data
func (user *User) PutArbitrary(data map[string]string) {
	if v, ok := data["email"]; ok {
		user.Email = &v
	}
}

// GetEmail Gets Email
func (user User) GetEmail() string {
	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	return email
}

// GetRecoverSelector Gets RecoverSelector
func (user User) GetRecoverSelector() string {
	return user.RecoverSelector
}

// GetRecoverVerifier Gets RecoverVerifier
func (user User) GetRecoverVerifier() string {
	return user.RecoverVerifier
}

// GetRecoverExpiry Gets RecoverExpiry
func (user User) GetRecoverExpiry() (expiry time.Time) {
	return user.RecoverExpiry
}

// PutEmail Puts Email
func (user *User) PutEmail(email string) {
	email = strings.ToLower(email)
	user.Email = &email
}

// PutRecoverSelector Puts RecoverSelector
func (user *User) PutRecoverSelector(selector string) {
	user.RecoverSelector = selector
}

// PutRecoverVerifier Puts RecoverVerifier
func (user *User) PutRecoverVerifier(verifier string) {
	user.RecoverVerifier = verifier
}

// PutRecoverExpiry Puts RecoverExpiry
func (user *User) PutRecoverExpiry(expiry time.Time) {
	user.RecoverExpiry = expiry
}
