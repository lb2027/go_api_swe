package main

import (
	"example/go_api_swe/database"

	"net/http"
)


func main() {
	
	mux := http.NewServeMux()

	// CRUD User
	mux.HandleFunc("/selectuser", database.Api_selectAllData)
	mux.HandleFunc("/adduser", database.API_add)
	mux.HandleFunc(("/deleteuser"), database.Api_deleteUser)
	mux.HandleFunc(("/updateuser"), database.Api_updateUser)

	// CRUD Produk
	mux.HandleFunc("/selectproduk", database.Api_selectAllProduk)
	mux.HandleFunc("/addproduk", database.API_addProduk)
	mux.HandleFunc("/deleteproduk", database.Api_deleteProduk)
	mux.HandleFunc("/updateproduk", database.Api_updateProduk)
	http.ListenAndServe(":5050", mux)

}
