package database

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	// "time" // Uncomment if you decide to use time.Time for date fields
)

// Gaji struct represents the structure of the 'gaji' table
type Gaji struct {
    ID              int     `json:"id"`
    StaffID         int     `json:"staff_id"`
    BulanGaji       string  `json:"bulan_gaji"`       // Expected format: YYYY-MM-DD
    GajiPerbulan    float64 `json:"gaji_perbulan"`
    TanggalTransfer string  `json:"tanggal_transfer"` // Expected format: YYYY-MM-DD
    Keterangan      *string `json:"keterangan,omitempty"` // Use pointer for nullable text field
}

// @Summary Add new salary record
// @Description Add a new salary record to the database
// @Tags Gaji
// @Accept json
// @Produce json
// @Param gaji body Gaji true "Salary data"
// @Success 201 {object} map[string]interface{} "message: Salary record added successfully, id: new_record_id"
// @Failure 400 {object} map[string]string "Error: Invalid JSON or Missing required fields"
// @Failure 401 {object} map[string]string "Error: Invalid token"
// @Failure 500 {object} map[string]string "Error: Internal server error"
// @Router /gaji [post]
func Api_AddGaji(w http.ResponseWriter, r *http.Request) {
    db := Koneksi()
    defer db.Close()
    enableCors(&w)

    token := r.Header.Get("token")
    if !isValidToken(token) {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    var gaji Gaji
    err := json.NewDecoder(r.Body).Decode(&gaji)
    if err != nil {
        http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
        return
    }

    // Basic validation
    if gaji.StaffID == 0 || gaji.BulanGaji == "" || gaji.GajiPerbulan == 0 || gaji.TanggalTransfer == "" {
        http.Error(w, "Missing required fields: staff_id, bulan_gaji, gaji_perbulan, tanggal_transfer are required", http.StatusBadRequest)
        return
    }

    var newID int
    query := `
        INSERT INTO gaji (staff_id, bulan_gaji, gaji_perbulan, tanggal_transfer, keterangan)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `
    err = db.QueryRow(query, gaji.StaffID, gaji.BulanGaji, gaji.GajiPerbulan, gaji.TanggalTransfer, gaji.Keterangan).Scan(&newID)
    if err != nil {
        log.Printf("Error inserting gaji record: %v\n", err)
        http.Error(w, "Failed to add salary record: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Salary record added successfully",
        "id":      newID,
    })
}

// @Summary Get all salary records
// @Description Get all salary records from the database
// @Tags Gaji
// @Accept json
// @Produce json
// @Success 200 {array} Gaji
// @Failure 401 {object} map[string]string "Error: Invalid token"
// @Failure 500 {object} map[string]string "Error: Internal server error"
// @Router /gaji [get]
func Api_GetAllGaji(w http.ResponseWriter, r *http.Request) {
    db := Koneksi()
    defer db.Close()
    enableCors(&w)

    token := r.Header.Get("token")
    if !isValidToken(token) {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    rows, err := db.Query("SELECT id, staff_id, bulan_gaji, gaji_perbulan, tanggal_transfer, keterangan FROM gaji ORDER BY bulan_gaji DESC, staff_id ASC")
    if err != nil {
        log.Printf("Error querying all gaji records: %v\n", err)
        http.Error(w, "Failed to retrieve salary records: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var allGaji []Gaji
    for rows.Next() {
        var gaji Gaji
        err := rows.Scan(&gaji.ID, &gaji.StaffID, &gaji.BulanGaji, &gaji.GajiPerbulan, &gaji.TanggalTransfer, &gaji.Keterangan)
        if err != nil {
            log.Printf("Error scanning gaji row: %v\n", err)
            http.Error(w, "Error processing salary records: "+err.Error(), http.StatusInternalServerError)
            return
        }
        allGaji = append(allGaji, gaji)
    }

    if err = rows.Err(); err != nil {
        log.Printf("Error after iterating gaji rows: %v\n", err)
        http.Error(w, "Error retrieving salary data: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(allGaji)
}

// @Summary Get salary record by ID
// @Description Get a specific salary record by its ID
// @Tags Gaji
// @Accept json
// @Produce json
// @Param id path int true "Gaji Record ID"
// @Success 200 {object} Gaji
// @Failure 400 {object} map[string]string "Error: Invalid ID"
// @Failure 401 {object} map[string]string "Error: Invalid token"
// @Failure 404 {object} map[string]string "Error: Salary record not found"
// @Failure 500 {object} map[string]string "Error: Internal server error"
// @Router /gaji/{id} [get]
func Api_GetGajiByID(w http.ResponseWriter, r *http.Request) {
    db := Koneksi()
    defer db.Close()
    enableCors(&w)

    token := r.Header.Get("token")
    if !isValidToken(token) {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    pathParts := strings.Split(r.URL.Path, "/")
    idStr := pathParts[len(pathParts)-1]
    id, err := strconv.Atoi(idStr)
    if err != nil || id <= 0 {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    var gaji Gaji
    query := "SELECT id, staff_id, bulan_gaji, gaji_perbulan, tanggal_transfer, keterangan FROM gaji WHERE id = $1"
    err = db.QueryRow(query, id).Scan(&gaji.ID, &gaji.StaffID, &gaji.BulanGaji, &gaji.GajiPerbulan, &gaji.TanggalTransfer, &gaji.Keterangan)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Salary record not found", http.StatusNotFound)
        } else {
            log.Printf("Error querying gaji by ID %d: %v\n", id, err)
            http.Error(w, "Failed to retrieve salary record: "+err.Error(), http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(gaji)
}

// @Summary Update salary record
// @Description Update an existing salary record in the database
// @Tags Gaji
// @Accept json
// @Produce json
// @Param gaji body Gaji true "Salary data to update. ID field must be provided."
// @Success 200 {object} map[string]string "message: Salary record updated successfully"
// @Failure 400 {object} map[string]string "Error: Invalid JSON, Missing required fields, or ID mismatch"
// @Failure 401 {object} map[string]string "Error: Invalid token"
// @Failure 404 {object} map[string]string "Error: Salary record not found"
// @Failure 500 {object} map[string]string "Error: Internal server error"
// @Router /gaji/{id} [put] // Or use PUT /gaji and get ID from body
func Api_UpdateGaji(w http.ResponseWriter, r *http.Request) {
    db := Koneksi()
    defer db.Close()
    enableCors(&w)

    token := r.Header.Get("token")
    if !isValidToken(token) {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    pathParts := strings.Split(r.URL.Path, "/")
    idStr := pathParts[len(pathParts)-1]
    idFromPath, err := strconv.Atoi(idStr)
    if err != nil || idFromPath <= 0 {
        http.Error(w, "Invalid ID in URL path", http.StatusBadRequest)
        return
    }

    var gaji Gaji
    err = json.NewDecoder(r.Body).Decode(&gaji)
    if err != nil {
        http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
        return
    }

    // Ensure ID from path matches ID in body if provided, or set it
    if gaji.ID == 0 {
        gaji.ID = idFromPath
    } else if gaji.ID != idFromPath {
        http.Error(w, "ID in path does not match ID in request body", http.StatusBadRequest)
        return
    }
    
    // Basic validation
    if gaji.StaffID == 0 || gaji.BulanGaji == "" || gaji.GajiPerbulan == 0 || gaji.TanggalTransfer == "" {
        http.Error(w, "Missing required fields: staff_id, bulan_gaji, gaji_perbulan, tanggal_transfer are required", http.StatusBadRequest)
        return
    }

    query := `
        UPDATE gaji
        SET staff_id = $1, bulan_gaji = $2, gaji_perbulan = $3, tanggal_transfer = $4, keterangan = $5
        WHERE id = $6
    `
    result, err := db.Exec(query, gaji.StaffID, gaji.BulanGaji, gaji.GajiPerbulan, gaji.TanggalTransfer, gaji.Keterangan, gaji.ID)
    if err != nil {
        log.Printf("Error updating gaji record ID %d: %v\n", gaji.ID, err)
        http.Error(w, "Failed to update salary record: "+err.Error(), http.StatusInternalServerError)
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        log.Printf("Error getting rows affected for gaji update ID %d: %v\n", gaji.ID, err)
        http.Error(w, "Failed to update salary record: "+err.Error(), http.StatusInternalServerError)
        return
    }

    if rowsAffected == 0 {
        http.Error(w, "Salary record not found or no changes made", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Salary record updated successfully"})
}

// @Summary Delete salary record
// @Description Delete a salary record by its ID
// @Tags Gaji
// @Accept json
// @Produce json
// @Param id path int true "Gaji Record ID"
// @Success 200 {object} map[string]string "message: Salary record deleted successfully"
// @Failure 400 {object} map[string]string "Error: Invalid ID"
// @Failure 401 {object} map[string]string "Error: Invalid token"
// @Failure 404 {object} map[string]string "Error: Salary record not found"
// @Failure 500 {object} map[string]string "Error: Internal server error"
// @Router /gaji/{id} [delete]
func Api_DeleteGaji(w http.ResponseWriter, r *http.Request) {
    db := Koneksi()
    defer db.Close()
    enableCors(&w)

    token := r.Header.Get("token")
    if !isValidToken(token) {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    pathParts := strings.Split(r.URL.Path, "/")
    idStr := pathParts[len(pathParts)-1]
    id, err := strconv.Atoi(idStr)
    if err != nil || id <= 0 {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    query := "DELETE FROM gaji WHERE id = $1"
    result, err := db.Exec(query, id)
    if err != nil {
        log.Printf("Error deleting gaji record ID %d: %v\n", id, err)
        http.Error(w, "Failed to delete salary record: "+err.Error(), http.StatusInternalServerError)
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        log.Printf("Error getting rows affected for gaji delete ID %d: %v\n", id, err)
        http.Error(w, "Failed to delete salary record: "+err.Error(), http.StatusInternalServerError)
        return
    }

    if rowsAffected == 0 {
        http.Error(w, "Salary record not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Salary record deleted successfully"})
}

// @Summary Get salary records by Staff ID
// @Description Get all salary records for a specific staff member
// @Tags Gaji
// @Accept json
// @Produce json
// @Param staff_id path int true "Staff ID"
// @Success 200 {array} Gaji
// @Failure 400 {object} map[string]string "Error: Invalid Staff ID"
// @Failure 401 {object} map[string]string "Error: Invalid token"
// @Failure 500 {object} map[string]string "Error: Internal server error"
// @Router /staff/{staff_id}/gaji [get]
func Api_GetGajiByStaffID(w http.ResponseWriter, r *http.Request) {
    db := Koneksi()
    defer db.Close()
    enableCors(&w)

    token := r.Header.Get("token")
    if !isValidToken(token) {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    pathParts := strings.Split(r.URL.Path, "/") // e.g. /staff/123/gaji
    if len(pathParts) < 4 { // expecting at least ["", "staff", "staff_id_value", "gaji"]
        http.Error(w, "Invalid URL path for staff ID", http.StatusBadRequest)
        return
    }
    staffIDStr := pathParts[len(pathParts)-2] // staff_id is the second to last part
    staffID, err := strconv.Atoi(staffIDStr)
    if err != nil || staffID <= 0 {
        http.Error(w, "Invalid Staff ID format", http.StatusBadRequest)
        return
    }

    rows, err := db.Query("SELECT id, staff_id, bulan_gaji, gaji_perbulan, tanggal_transfer, keterangan FROM gaji WHERE staff_id = $1 ORDER BY bulan_gaji DESC", staffID)
    if err != nil {
        log.Printf("Error querying gaji records for staff ID %d: %v\n", staffID, err)
        http.Error(w, "Failed to retrieve salary records for staff: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var staffGajiRecords []Gaji
    for rows.Next() {
        var gaji Gaji
        err := rows.Scan(&gaji.ID, &gaji.StaffID, &gaji.BulanGaji, &gaji.GajiPerbulan, &gaji.TanggalTransfer, &gaji.Keterangan)
        if err != nil {
            log.Printf("Error scanning gaji row for staff ID %d: %v\n", staffID, err)
            http.Error(w, "Error processing salary records for staff: "+err.Error(), http.StatusInternalServerError)
            return
        }
        staffGajiRecords = append(staffGajiRecords, gaji)
    }

    if err = rows.Err(); err != nil {
        log.Printf("Error after iterating gaji rows for staff ID %d: %v\n", staffID, err)
        http.Error(w, "Error retrieving salary data for staff: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(staffGajiRecords)
}