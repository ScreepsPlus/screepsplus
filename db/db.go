package db

import (
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // mysql driver
	"github.com/screepsplus/screepsplus/models"
)

// DB main DB instance
var DB *gorm.DB

var defaults = map[string]string{
	"MYSQL_HOST": "localhost",
	"MYSQL_PORT": "3306",
	"MYSQL_USER": "root",
	"MYSQL_PASS": "",
	"MYSQL_DB":   "screepsplus",
}

// Init loads and preps the database
func init() {
	mapper := func(key string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return defaults[key]
	}
	connStr := os.Expand("${MYSQL_USER}:${MYSQL_PASS}@tcp(${MYSQL_HOST}:${MYSQL_PORT})/${MYSQL_DB}?charset=utf8&parseTime=True&loc=Local", mapper)
	db, err := gorm.Open("mysql", connStr)
	defer db.Close()
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&models.User{})
	var firstUser = models.User{Username: "admin", Email: "admin@example.com"}
	firstUser.SetNewPassword("admin")
	firstUser.Active = true
	db.Create(&firstUser)
	DB = db
}
