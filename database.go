package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


var DB *gorm.DB
func ConnectionDatabase() {
	database, err := gorm.Open(postgres.Open("frozen_food.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal terhubung ke database!")
	}
	DB = database
}

