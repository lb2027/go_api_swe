package main

import (
	"example/go_api_swe/database"
	"net/http"

	_ "example/go_api_swe/docs" // Swagger docs (hasil dari swag init)

	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger" // Swagger UI
)

// @securityDefinitions.apikey BearerAuth
// @in header
// @name token

func main() {

	mux := http.NewServeMux()

	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	mux.HandleFunc("/login", database.API_generateJWT) 
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
	mux.Handle("/soldproduk", database.MiddleWare(database.Api_soldProduk))

	// Transaksi
	mux.Handle("/selecttransaksi", database.MiddleWare(database.Api_selectAllTransaksi))
	mux.Handle("/addtransaksi", database.MiddleWare(database.Api_addTransaksi))
	mux.Handle("/displayhistory", database.MiddleWare(database.GetTransactionHistory))
	mux.Handle("/dailysales", database.MiddleWare(database.GetDailySales))
	mux.Handle("/weeklysales", database.MiddleWare(database.GetWeeklySales))
	mux.Handle("/monthlysales", database.MiddleWare(database.GetMonthlyRevenue))
	mux.Handle("/inventorystatus", database.MiddleWare(database.GetInventoryStatus))

	// Staff
	// Staff management endpoints
	mux.Handle("/staff", database.MiddleWare(database.Api_getAllStaff))
	mux.Handle("/staff/", database.MiddleWare(database.Api_getStaffByID))
	mux.Handle("/addstaff", database.MiddleWare(database.Api_addStaff))
	mux.Handle("/updatestaff", database.MiddleWare(database.Api_updateStaff))
	mux.Handle("/deletestaff", database.MiddleWare(database.Api_deleteStaff))
	mux.Handle("/bulkdeletestaff", database.MiddleWare(database.Api_bulkDeleteStaff))
	mux.Handle("/bulkupdatestaffstatus", database.MiddleWare(database.Api_bulkUpdateStaffStatus))
	mux.Handle("/searchstaff", database.MiddleWare(database.Api_searchStaff))
	mux.Handle("/staffstats", database.MiddleWare(database.Api_getStaffStats))
	mux.Handle("/makemestaff", database.MiddleWare(database.Api_getStaff))

	// haruse keluar lo ini
	// CORS setup
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "token", "Authorization"},
	})

	handler := c.Handler(mux)

	// Start the server
	print("Server running on port 5050")
	http.ListenAndServe(":5050", handler)
}
