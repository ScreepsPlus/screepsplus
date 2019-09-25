package auth

import (
	"context"
	"log"

	"github.com/screepsplus/screepsplus/db"
	"github.com/screepsplus/screepsplus/models"
	"github.com/volatiletech/authboss"
)

// ServerStorer is a db adapter for authboss
type ServerStorer struct{}

var testUsers = []authboss.User{}

// New creates a new user
func (s ServerStorer) New(ctx context.Context) authboss.User {
	return models.NewUser()
}

// Create saves a new user
func (s ServerStorer) Create(ctx context.Context, user authboss.User) error {
	u := user.(*models.User)
	log.Printf("Create %s %v", user.GetPID(), u)
	if err := db.DB().Create(&u).Error; err != nil {
		return authboss.ErrUserFound
	}
	return nil
}

// Load loads a user
func (s ServerStorer) Load(ctx context.Context, key string) (authboss.User, error) {
	log.Printf("Load %s", key)
	user := models.User{}
	if db.DB().Where(&models.User{Username: key}).First(&user).RecordNotFound() {
		return nil, authboss.ErrUserNotFound
	}
	return &user, nil
}

// Save saves a user
func (s ServerStorer) Save(ctx context.Context, user authboss.User) error {
	u := user.(*models.User)
	log.Printf("Save %s %v", user.GetPID(), user)
	if db.DB().Where(&models.User{Username: u.GetPID()}).First(&models.User{}).RecordNotFound() {
		return authboss.ErrUserNotFound
	}
	if err := db.DB().Save(&u).Error; err != nil {
		log.Printf("Err while saving: %v", err)
		return err
	}
	return nil
}

// LoadByRecoverSelector loads a user by recovery selector
func (s ServerStorer) LoadByRecoverSelector(ctx context.Context, selector string) (authboss.RecoverableUser, error) {
	log.Printf("LoadByRecoverSelector %s", selector)
	user := models.User{}
	if db.DB().Where(&models.User{RecoverSelector: selector}).First(&user).RecordNotFound() {
		return nil, authboss.ErrUserNotFound
	}
	return &user, nil
}
