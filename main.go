package main

import (
	"example/go_api_swe/models"
	"fmt"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
)
var transaksi = []models.Transaksi{
	{TransaksiID: 1, NamaProduk: "Indomie", Harga: 2500, JumlahTerjual: 10, TotalHarga: 25000, Tanggal: time.Now()},
	{TransaksiID: 2, NamaProduk: "Mie Sedap", Harga: 3000, JumlahTerjual: 5, TotalHarga: 15000, Tanggal: time.Now()},
}


func getTransaksi(context *gin.Context) {
	fmt.Println(transaksi)
	context.IndentedJSON(http.StatusOK, transaksi)
}

func getProduk(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, models.Produk{})
}

func addProduk(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, models.Produk{})
}

func updateProduk(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, models.Produk{})
}



func main() {
	router := gin.Default()
	router.GET("/transaksi", getTransaksi)
	router.GET("/produk", getProduk)
	router.POST("/produk", addProduk)
	router.PUT("/produk", updateProduk)
	router.Run("localhost:8080")
}
