package main

import (
	"example/go_api_swe/database"

	"net/http"

	"github.com/rs/cors"
)



func main() {
	mux := http.NewServeMux()

	// CRUD User
	mux.HandleFunc("/selectuser", database.Api_selectAllData) // sudah
	mux.HandleFunc("/adduser", database.API_add) // sudah
	mux.HandleFunc(("/deleteuser"), database.Api_deleteUser)
	mux.HandleFunc(("/updateuser"), database.Api_updateUser)

	// CRUD Produk
	mux.HandleFunc("/selectproduk", database.Api_selectAllProduk) // sudah
	mux.HandleFunc("/addproduk", database.API_addProduk) // sudah
	mux.HandleFunc("/deleteproduk", database.Api_deleteProduk)
	mux.HandleFunc("/updateproduk", database.Api_updateProduk)

	// GET Transaksi
	mux.HandleFunc("/selecttransaksi", database.Api_selectAllTransaksi)
	mux.HandleFunc("/addtransaksi", database.Api_addTransaksi)

	// mux = http.NewServeMux()
	// mux.HandleFunc("/selectuser", database.Api_selectAllData)
	// mux.HandleFunc("/adduser",database.API_add)
	// mux.HandleFunc("/jwt",database.API_generateJWT)
	// mux.HandleFunc("/jwt", database.MiddleWare(database.API_generateJWT))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},	
		// AllowedCredentials: true,
	})

	handler := c.Handler(mux)
	http.ListenAndServe(":5050", handler)
}

