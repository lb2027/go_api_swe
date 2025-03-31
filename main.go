package main

import (
	"example/go_api_swe/database"
	"net/http"

	"github.com/rs/cors"
)

func main() {
	mux := http.NewServeMux()

	// Public routes (no authentication required)
	mux.HandleFunc("/login", database.API_generateJWT) // Changed from /jwt to /login

	// Protected routes (require authentication)
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

	// GET Transaksi
	mux.Handle("/selecttransaksi", database.MiddleWare(database.Api_selectAllTransaksi))
	mux.Handle("/addtransaksi", database.MiddleWare(database.Api_addTransaksi))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Content-Type", "token"}, // Add "token" to allowed headers
	})

	handler := c.Handler(mux)
	http.ListenAndServe(":5050", handler)
}

