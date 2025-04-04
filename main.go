package main

import (
	"example/go_api_swe/database"
	"net/http"

	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger" // Swagger UI handler
	_ "example/go_api_swe/docs"                   // Swagger docs (hasil dari swag init)
)

// @title Frozen Food
// @version 1.0
// @description Sistem REST API untuk manajemen stok produk frozen food, pencatatan transaksi, serta autentikasi pengguna menggunakan JWT.
// @host localhost:5050
// @BasePath /

// main function
func main() {
	mux := http.NewServeMux()

	// Swagger route
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Public routes
	mux.HandleFunc("/login", database.API_generateJWT)

	// Protected routes (pakai middleware JWT)
	// CRUD User
	mux.Handle("/selectuser", database.MiddleWare(database.Api_selectAllData))
	mux.Handle("/adduser", database.MiddleWare(database.API_add))
	mux.Handle("/deleteuser", database.MiddleWare(database.Api_deleteUser))
	mux.Handle("/updateuser", database.MiddleWare(database.Api_updateUser))

	// CRUD Produk
	mux.Handle("/selectproduk", database.MiddleWare(database.Api_selectAllProduk))
	mux.Handle("/addproduk", database.MiddleWare(database.API_addProduk))
	mux.Handle("/deleteproduk", database.MiddleWare(database.Api_deleteProduk))
	mux.Handle("/updateproduk", database.MiddleWare(database.Api_updateProduk))

	// Transaksi
	mux.Handle("/selecttransaksi", database.MiddleWare(database.Api_selectAllTransaksi))
	mux.Handle("/addtransaksi", database.MiddleWare(database.Api_addTransaksi))

	// CORS setup
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Content-Type", "token"},
	})

	handler := c.Handler(mux)

	// Start server
	http.ListenAndServe(":5050", handler)
}
