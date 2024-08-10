package database

import (
	"log"
	"os"

	"github.com/karan-singh-17/Quick-Mail/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("database_url")
	connection, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	DB = connection
	connection.AutoMigrate(&models.User{}, &models.Group{})
	log.Println("Database connection successful")
}
