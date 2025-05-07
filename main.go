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

	// Connect to the database
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	mux.HandleFunc("/login", database.API_generateJWT)
	mux.HandleFunc("/register", database.API_register)

	// Protected routes (pakai middleware JWT)
	// CRUD User (4 endpoints)
	mux.Handle("/selectuser", database.MiddleWare(database.Api_selectAllData))
	mux.Handle("/adduser", database.MiddleWare(database.API_add))
	mux.Handle("/deleteuser", database.MiddleWare(database.Api_deleteUser))
	mux.Handle("/updateuser", database.MiddleWare(database.Api_updateUser))

	// CRUD Produk (6 endpoints)
	mux.Handle("/selectproduk", database.MiddleWare(database.Api_selectAllProduk))
	mux.Handle("/addproduk", database.MiddleWare(database.API_addProduk))
	mux.Handle("/deleteproduk", database.MiddleWare(database.Api_deleteProduk))
	mux.Handle("/updateproduk", database.MiddleWare(database.Api_updateProduk))
	mux.Handle("/selectProdukById", database.MiddleWare(database.Api_selectProdukById))
	mux.Handle("/soldproduk", database.MiddleWare(database.Api_soldProduk))

	// Transaksi endpoints (7 endpoints)
	mux.Handle("/selecttransaksi", database.MiddleWare(database.Api_selectAllTransaksi))
	mux.Handle("/addtransaksi", database.MiddleWare(database.Api_addTransaksi))
	mux.Handle("/displayhistory", database.MiddleWare(database.GetTransactionHistory))
	mux.Handle("/dailysales", database.MiddleWare(database.GetDailySales))
	mux.Handle("/weeklysales", database.MiddleWare(database.GetWeeklySales))
	mux.Handle("/monthlysales", database.MiddleWare(database.GetMonthlyRevenue))
	mux.Handle("/inventorystatus", database.MiddleWare(database.GetInventoryStatus))

	// Absensi
	mux.Handle("/addabsensi", database.MiddleWare(database.Api_addAbsensi)) // POST

	// Staff management endpoints (11 endpoints)
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
	mux.Handle("/getabsensi", database.MiddleWare(database.Api_getAttendanceLogs))

	// Gaji management endpoints (4 endpoints)
	mux.Handle("/getgaji", database.MiddleWare(database.Api_GetAllGaji)) // GET
	mux.Handle("/addgaji", database.MiddleWare(database.Api_AddGaji)) // POST
	mux.Handle("/updategaji", database.MiddleWare(database.Api_UpdateGaji)) // PUT
	mux.Handle("/deletegaji", database.MiddleWare(database.Api_DeleteGaji)) // DELETE



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
