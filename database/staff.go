package database

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Staff struct {
	Staff_id      int    `json:"id"`
	Username      string `json:"nama"`
	NomorHP       string `json:"no_hp"`
	Alamat        string `json:"alamat"`
	Email         string `json:"email"`
	StatusKerja   string `json:"status_kerja"`
	User_id       int    `json:"user_id"`
	Tanggal_lahir string `json:"tanggal_lahir"` // Added field for DOB
}

// @Summary Make me staff
// @Description Make me staff so I can get access to the database
// @Tags Staff
// @Accept json
// @Produce json
// @Param staff body Staff true "Data staff"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /makemestaff [post]
func Api_getStaff(w http.ResponseWriter, r *http.Request) {
	db := Koneksi()
	defer db.Close()
	// Enable CORS for the response
	enableCors(&w)
	var staff Staff
	err := json.NewDecoder(r.Body).Decode(&staff)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Check if the user is already a staff member
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM staff WHERE username = ?", staff.Username).Scan(&count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "User is already a staff member", http.StatusBadRequest)
		return
	}
	// Insert the new staff member into the database
	_, err = db.Exec("INSERT INTO staff (username) VALUES (?)", staff.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "User is now a staff member"}`))
}

// @Summary Get all staff
// @Description Get all staff from the database
// @Tags Staff
// @Accept json
// @Produce json
// @Success 200 {array} Staff
// @Failure 500 {object} map[string]string
// @Router /staff [get]
func Api_getAllStaff(w http.ResponseWriter, r *http.Request) {
	db := Koneksi()
	defer db.Close()
	enableCors(&w)

	// Check token
	token := r.Header.Get("token")
	if !isValidToken(token) {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get all staff
	rows, err := db.Query("SELECT id, nama, no_hp, alamat, email, status_kerja, user_id, tanggal_lahir FROM staff ORDER BY nama")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Parse rows into Staff struct
	var staffs []Staff
	for rows.Next() {
		var staff Staff
		var userID *int                        // Nullable field
		var email, alamat, dateOfBirth *string // Nullable fields

		err := rows.Scan(&staff.Staff_id, &staff.Username, &staff.NomorHP, &alamat, &email, &staff.StatusKerja, &userID, &dateOfBirth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Handle nullable fields
		if userID != nil {
			staff.User_id = *userID
		}
		if email != nil {
			staff.Email = *email
		}
		if alamat != nil {
			staff.Alamat = *alamat
		}
		if dateOfBirth != nil {
			staff.Tanggal_lahir = *dateOfBirth
		}

		staffs = append(staffs, staff)
	}

	// Return the staffs as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(staffs)
}

// @Summary Get staff by ID
// @Description Get staff by ID from the database
// @Tags Staff
// @Accept json
// @Produce json
// @Param id path int true "Staff ID"
// @Success 200 {object} Staff
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /staff/{id} [get]
func Api_getStaffByID(w http.ResponseWriter, r *http.Request) {
	db := Koneksi()
	defer db.Close()
	enableCors(&w)

	// Check token
	token := r.Header.Get("token")
	if !isValidToken(token) {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get ID from URL path
	urlParts := strings.Split(r.URL.Path, "/")
	if len(urlParts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(urlParts[len(urlParts)-1])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Get staff by ID
	row := db.QueryRow("SELECT id, nama, no_hp, alamat, email, status_kerja, user_id, tanggal_lahir FROM staff WHERE id = $1", id)

	var staff Staff
	var userID *int                        // Nullable field
	var email, alamat, dateOfBirth *string // Nullable fields

	err = row.Scan(&staff.Staff_id, &staff.Username, &staff.NomorHP, &alamat, &email, &staff.StatusKerja, &userID, &dateOfBirth)
	if err != nil {
		http.Error(w, "Staff not found", http.StatusNotFound)
		return
	}

	// Handle nullable fields
	if userID != nil {
		staff.User_id = *userID
	}
	if email != nil {
		staff.Email = *email
	}
	if alamat != nil {
		staff.Alamat = *alamat
	}
	if dateOfBirth != nil {
		staff.Tanggal_lahir = *dateOfBirth
	}

	// Return the staff as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(staff)
}

// @Summary Add new staff
// @Description Add a new staff member to the database
// @Tags Staff
// @Accept json
// @Produce json
// @Param staff body Staff true "Staff data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /addstaff [post]
func Api_addStaff(w http.ResponseWriter, r *http.Request) {
	db := Koneksi()
	defer db.Close()
	enableCors(&w)

	// Check token
	token := r.Header.Get("token")
	if !isValidToken(token) {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var staff Staff
	err := json.NewDecoder(r.Body).Decode(&staff)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if staff.Username == "" || staff.NomorHP == "" || staff.StatusKerja == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Insert new staff and return the ID (PostgreSQL way with RETURNING clause)
	var staffID int
	var query string
	var args []interface{}

	if staff.User_id > 0 {
		if staff.Tanggal_lahir != "" {
			query = `INSERT INTO staff (nama, no_hp, alamat, email, status_kerja, user_id, tanggal_lahir) 
					VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
			args = []interface{}{staff.Username, staff.NomorHP, staff.Alamat, staff.Email,
				staff.StatusKerja, staff.User_id, staff.Tanggal_lahir}
		} else {
			query = `INSERT INTO staff (nama, no_hp, alamat, email, status_kerja, user_id) 
					VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
			args = []interface{}{staff.Username, staff.NomorHP, staff.Alamat, staff.Email,
				staff.StatusKerja, staff.User_id}
		}
	} else {
		if staff.Tanggal_lahir != "" {
			query = `INSERT INTO staff (nama, no_hp, alamat, email, status_kerja, tanggal_lahir) 
					VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
			args = []interface{}{staff.Username, staff.NomorHP, staff.Alamat, staff.Email,
				staff.StatusKerja, staff.Tanggal_lahir}
		} else {
			query = `INSERT INTO staff (nama, no_hp, alamat, email, status_kerja) 
					VALUES ($1, $2, $3, $4, $5) RETURNING id`
			args = []interface{}{staff.Username, staff.NomorHP, staff.Alamat, staff.Email,
				staff.StatusKerja}
		}
	}

	err = db.QueryRow(query, args...).Scan(&staffID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Staff added successfully",
		"id":      staffID,
	})
}

// @Summary Update staff
// @Description Update an existing staff member in the database
// @Tags Staff
// @Accept json
// @Produce json
// @Param staff body Staff true "Staff data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /updatestaff [put]
func Api_updateStaff(w http.ResponseWriter, r *http.Request) {
	db := Koneksi()
	defer db.Close()
	enableCors(&w)

	// Check token
	token := r.Header.Get("token")
	if !isValidToken(token) {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var staff Staff
	err := json.NewDecoder(r.Body).Decode(&staff)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if staff.Staff_id == 0 || staff.Username == "" || staff.NomorHP == "" || staff.StatusKerja == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Check if staff exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM staff WHERE id = $1", staff.Staff_id).Scan(&count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count == 0 {
		http.Error(w, "Staff not found", http.StatusNotFound)
		return
	}

	// Update staff - Fixed for PostgreSQL syntax
	var result sql.Result
	if staff.User_id > 0 {
		if staff.Tanggal_lahir != "" {
			result, err = db.Exec(
				"UPDATE staff SET nama = $1, no_hp = $2, alamat = $3, email = $4, status_kerja = $5, user_id = $6, tanggal_lahir = $7 WHERE id = $8",
				staff.Username, staff.NomorHP, staff.Alamat, staff.Email, staff.StatusKerja, staff.User_id, staff.Tanggal_lahir, staff.Staff_id)
		} else {
			result, err = db.Exec(
				"UPDATE staff SET nama = $1, no_hp = $2, alamat = $3, email = $4, status_kerja = $5, user_id = $6, tanggal_lahir = NULL WHERE id = $7",
				staff.Username, staff.NomorHP, staff.Alamat, staff.Email, staff.StatusKerja, staff.User_id, staff.Staff_id)
		}
	} else {
		if staff.Tanggal_lahir != "" {
			result, err = db.Exec(
				"UPDATE staff SET nama = $1, no_hp = $2, alamat = $3, email = $4, status_kerja = $5, user_id = NULL, tanggal_lahir = $6 WHERE id = $7",
				staff.Username, staff.NomorHP, staff.Alamat, staff.Email, staff.StatusKerja, staff.Tanggal_lahir, staff.Staff_id)
		} else {
			result, err = db.Exec(
				"UPDATE staff SET nama = $1, no_hp = $2, alamat = $3, email = $4, status_kerja = $5, user_id = NULL, tanggal_lahir = NULL WHERE id = $6",
				staff.Username, staff.NomorHP, staff.Alamat, staff.Email, staff.StatusKerja, staff.Staff_id)
		}
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "No changes made", http.StatusBadRequest)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Staff updated successfully",
	})
}

// @Summary Delete staff
// @Description Delete a staff member from the database
// @Tags Staff
// @Accept json
// @Produce json
// @Param id body map[string]int true "Staff ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /deletestaff [delete]
func Api_deleteStaff(w http.ResponseWriter, r *http.Request) {
	db := Koneksi()
	defer db.Close()
	enableCors(&w)

	// Check token
	token := r.Header.Get("token")
	if !isValidToken(token) {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var requestBody map[string]int
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, exists := requestBody["id"]
	if !exists || id <= 0 {
		http.Error(w, "Invalid or missing ID", http.StatusBadRequest)
		return
	}

	// Check if staff exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM staff WHERE id = $1", id).Scan(&count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count == 0 {
		http.Error(w, "Staff not found", http.StatusNotFound)
		return
	}

	// Delete staff
	result, err := db.Exec("DELETE FROM staff WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Staff not deleted", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Staff deleted successfully",
	})
}

// @Summary Bulk delete staff
// @Description Delete multiple staff members from the database
// @Tags Staff
// @Accept json
// @Produce json
// @Param ids body map[string][]int true "Staff IDs"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /bulkdeletestaff [delete]
func Api_bulkDeleteStaff(w http.ResponseWriter, r *http.Request) {
	db := Koneksi()
	defer db.Close()
	enableCors(&w)

	// Check token
	token := r.Header.Get("token")
	if !isValidToken(token) {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var requestBody map[string][]int
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ids, exists := requestBody["ids"]
	if !exists || len(ids) == 0 {
		http.Error(w, "Invalid or missing IDs", http.StatusBadRequest)
		return
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build a query with numbered placeholders
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}

	query := "DELETE FROM staff WHERE id IN (" + strings.Join(placeholders, ",") + ")"
	result, err := tx.Exec(query, args...)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Staff deleted successfully",
		"deletedCount": rowsAffected,
	})
}

// @Summary Bulk update staff status
// @Description Update status for multiple staff members
// @Tags Staff
// @Accept json
// @Produce json
// @Param data body map[string]interface{} true "Staff IDs and status"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /bulkupdatestaffstatus [put]
func Api_bulkUpdateStaffStatus(w http.ResponseWriter, r *http.Request) {
	db := Koneksi()
	defer db.Close()
	enableCors(&w)

	// Check token
	token := r.Header.Get("token")
	if !isValidToken(token) {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var requestBody struct {
		Ids         []int  `json:"ids"`
		StatusKerja string `json:"status_kerja"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(requestBody.Ids) == 0 || requestBody.StatusKerja == "" {
		http.Error(w, "Missing IDs or status", http.StatusBadRequest)
		return
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update staff status - Fixed for PostgreSQL
	placeholders := make([]string, len(requestBody.Ids))
	args := make([]interface{}, len(requestBody.Ids)+1)
	args[0] = requestBody.StatusKerja

	for i, id := range requestBody.Ids {
		placeholders[i] = "$" + strconv.Itoa(i+2) // Start with $2, $3, etc.
		args[i+1] = id
	}

	query := "UPDATE staff SET status_kerja = $1 WHERE id IN (" + strings.Join(placeholders, ",") + ")"
	result, err := tx.Exec(query, args...)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Staff status updated successfully",
		"updatedCount": rowsAffected,
	})
}

// @Summary Search staff
// @Description Search for staff by name, email, phone, or status
// @Tags Staff
// @Accept json
// @Produce json
// @Param query query string true "Search query"
// @Success 200 {array} Staff
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /searchstaff [get]
func Api_searchStaff(w http.ResponseWriter, r *http.Request) {
	db := Koneksi()
	defer db.Close()
	enableCors(&w)

	// Check token
	token := r.Header.Get("token")
	if !isValidToken(token) {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get query parameter
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Missing search query", http.StatusBadRequest)
		return
	}

	// Search staff by name, email, phone, or status - Fixed for PostgreSQL
	rows, err := db.Query(`
        SELECT id, nama, no_hp, alamat, email, status_kerja, user_id, tanggal_lahir 
        FROM staff 
        WHERE nama ILIKE $1 OR email ILIKE $2 OR no_hp ILIKE $3 OR status_kerja ILIKE $4 
        ORDER BY nama
    `, "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Parse rows into Staff struct
	var staffs []Staff
	for rows.Next() {
		var staff Staff
		var userID *int                        // Nullable field
		var email, alamat, dateOfBirth *string // Nullable fields

		err := rows.Scan(&staff.Staff_id, &staff.Username, &staff.NomorHP, &alamat, &email, &staff.StatusKerja, &userID, &dateOfBirth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Handle nullable fields
		if userID != nil {
			staff.User_id = *userID
		}
		if email != nil {
			staff.Email = *email
		}
		if alamat != nil {
			staff.Alamat = *alamat
		}
		if dateOfBirth != nil {
			staff.Tanggal_lahir = *dateOfBirth
		}

		staffs = append(staffs, staff)
	}

	// Return the staffs as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(staffs)
}

// @Summary Get staff statistics
// @Description Get statistics about staff members
// @Tags Staff
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /staffstats [get]
func Api_getStaffStats(w http.ResponseWriter, r *http.Request) {
	db := Koneksi()
	defer db.Close()
	enableCors(&w)

	// Check token
	token := r.Header.Get("token")
	if !isValidToken(token) {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get total staff count
	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM staff").Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get count by status
	rows, err := db.Query("SELECT status_kerja, COUNT(*) FROM staff GROUP BY status_kerja")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Parse rows into map
	statusCount := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		err := rows.Scan(&status, &count)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		statusCount[status] = count
	}

	// Return the stats as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total":       total,
		"statusCount": statusCount,
	})
}

// Authentication helper function
func isValidToken(token string) bool {
	if token == "" {
		return false
	}

	// In a real application, you would validate the token
	// For simplicity, we're just checking if it's not empty
	return true
}

// @Summary Get staff statistics
// @Description Get absensi Staff
// @Tags Staff
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /staffstats [get]
func Api_addAbsensi(w http.ResponseWriter, r *http.Request) {
	db := Koneksi()
	defer db.Close()
	enableCors(&w)

	// Check token
	token := r.Header.Get("token")
	if !isValidToken(token) {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Latitude  string `json:"latitude"`
		Longitude string `json:"longitude"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO absensi (latitude, longitude) VALUES ($1, $2)", data.Latitude, data.Longitude)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Absensi berhasil ditambahkan"})
}
