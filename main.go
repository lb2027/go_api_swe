package main

// go mod init go-api
// go get -u github.com/gin-gonic/gin
// go get -u gorm.io/gorm
// go get -u gorm.io/driver/postgres
import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"errors"
	"net/http"

)

func main() {
	dsn := "host=localhost user=postgres password=;';' dbname=todo port=5432 sslmode=disable TimeZone=Asia/Jakarta" // Ganti dengan detail koneksi Anda
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}


	router := gin.Default()

	

	router.Run(":8080")
}
