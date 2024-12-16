package db

import (
	"log"

	"mobilehub-order/pkg/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func Init(url string) Handler {
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}

	db.AutoMigrate(&models.Order{})
	db.AutoMigrate(&models.OrderItem{})

	return Handler{db}
}
