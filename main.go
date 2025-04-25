package main

import (
	"example/go_api_swe/database"
	"net/http"

	_ "example/go_api_swe/docs" // Swagger docs (hasil dari swag init)

	_ "example/go_api_swe/docs" // Swagger docs (hasil dari swag init)

	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	// Public routes (no authentication required)
	mux.HandleFunc("/login", database.API_generateJWT) // Changed from /jwt to /login
	mux.HandleFunc("/register", database.API_register)

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
	mux.Handle("/selectProdukById", database.MiddleWare(database.Api_selectProdukById))


	// Transaksi
	mux.Handle("/selecttransaksi", database.MiddleWare(database.Api_selectAllTransaksi))
	mux.Handle("/addtransaksi", database.MiddleWare(database.Api_addTransaksi))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Content-Type", "token"},
	})

	handler := c.Handler(mux)

	// Start server
	http.ListenAndServe(":5050", handler)
}
