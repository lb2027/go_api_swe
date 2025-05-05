package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Produk struct {
	ProdukID  int     `json:"produk_id"`
	Nama      string  `json:"nama"`
	Stok      int     `json:"stok"`
	Harga     float64 `json:"harga"`
	HargaBeli float64 `json:"harga_beli"`
	Foto      string  `json:"foto"`
	Supplier  string  `json:"supplier"`
}

type SoldItem struct {
	ProdukID   int     `json:"produk_id"`
	NamaProduk string  `json:"nama_produk"`
	Harga      float64 `json:"harga"`
	StokKeluar int     `json:"stok_keluar"`
}

func select_allProduk() []Produk {
	db := Koneksi()
	var produk = Produk{}
	var produkArray []Produk

	statement, err := db.Query("SELECT * FROM produk")

	if err != nil {
		panic(err.Error())
	} else {
		for statement.Next() {
			err = statement.Scan(&produk.ProdukID, &produk.Nama, &produk.Stok, &produk.Harga, &produk.HargaBeli, &produk.Foto, &produk.Supplier)
			if err != nil {
				panic(err.Error())
			} else {
				produkArray = append(produkArray, Produk{ProdukID: produk.ProdukID, Nama: produk.Nama, Stok: produk.Stok, Harga: produk.Harga, HargaBeli: produk.HargaBeli, Foto: produk.Foto, Supplier: produk.Supplier})
			}
		}
	}
	return produkArray
}

func getProdukById(id string) (Produk, error) {
	db := Koneksi()
	defer db.Close()

	var produk Produk

	statement, err := db.Query("SELECT produk_id, nama, stok, harga, harga_beli, foto, supplier FROM produk WHERE produk_id = $1", id)

	if err != nil {
		return Produk{}, err
	}

	defer statement.Close()

	for statement.Next() {
		err = statement.Scan(&produk.ProdukID, &produk.Nama, &produk.Stok, &produk.Harga, &produk.HargaBeli, &produk.Foto, &produk.Supplier)
		if err != nil {
			return Produk{}, err
		} else {
			return produk, nil
		}
	}

	return Produk{}, sql.ErrNoRows
}

func addDataProduk(nama string, stok int, harga float64, harga_beli float64, foto string, supplier string) {
	db := Koneksi()
	statement, err := db.Prepare("INSERT INTO produk (nama, stok, harga, harga_beli, foto, supplier) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		panic(err.Error())
	} else {
		statement.Exec(nama, stok, harga, harga_beli, foto, supplier)
	}
}

func updateDataProduk(id int, nama string, stok int, harga float64, harga_beli float64, foto string, supplier string) {
	db := Koneksi()
	defer db.Close()

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		panic(err.Error())
	}

	// Get the original stock value
	var originalStok int
	err = tx.QueryRow("SELECT stok FROM produk WHERE produk_id = $1", id).Scan(&originalStok)
	if err != nil {
		tx.Rollback()
		panic(err.Error())
	}

	// Calculate the stock difference
	stockDifference := stok - originalStok

	// Update the produk table
	_, err = tx.Exec("UPDATE produk SET nama = $1, stok = $2, harga = $3, harga_beli = $4, foto = $5, supplier = $6 WHERE produk_id = $7", nama, stok, harga, harga_beli, foto, supplier, id)
	if err != nil {
		tx.Rollback()
		panic(err.Error())
	}

	// Insert a record into the stok_masuk table
	_, err = tx.Exec("INSERT INTO stok_masuk (produk_id, jumlah) VALUES ($1, $2)", id, stockDifference)
	if err != nil {
		tx.Rollback()
		panic(err.Error())
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		panic(err.Error())
	}
}

func deleteDataProduk(id string) {
	db := Koneksi()
	statement, err := db.Prepare("DELETE FROM produk WHERE produk_id = $1")
	if err != nil {
		panic(err.Error())
	} else {
		statement.Exec(id)
	}
}

func updateStockAndRecordSale(productID int, quantitySold int, transaksiID int) error {
	db := Koneksi()
	defer db.Close()

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Update the produk table
	result, err := tx.Exec("UPDATE produk SET stok = stok - $1 WHERE produk_id = $2 AND stok >= $1", quantitySold, productID) // Add check to prevent negative stock
	if err != nil {
		tx.Rollback()
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("no rows updated, either product ID %d not found or not enough stock", productID)
	}

	// Insert a record into the stok_keluar table
	currentTime := time.Now()
	_, err = tx.Exec("INSERT INTO stok_keluar (produk_id, jumlah, tanggal) VALUES ($1, $2, $3)", productID, quantitySold, currentTime)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Get product details
	var namaProduk string
	var harga float64
	err = tx.QueryRow("SELECT nama, harga FROM produk WHERE produk_id = $1", productID).Scan(&namaProduk, &harga)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Calculate total harga
	totalHarga := harga * float64(quantitySold)

	// Insert a record into the transaksi table
	_, err = tx.Exec("INSERT INTO transaksi (nama_produk, harga, jumlah_terjual, total_harga, tanggal) VALUES ($1, $2, $3, $4, $5)", namaProduk, harga, quantitySold, totalHarga, time.Now())
	if err != nil {
		fmt.Printf("Error inserting into transaksi table: %v\n", err)
		tx.Rollback()
		return fmt.Errorf("failed to insert into transaksi table: %w", err) // Wrap the error
	}

	// Insert into transaksi_detail using the provided transaksiID
	_, err = tx.Exec(
		"INSERT INTO transaksi_detail (transaksi_id, nama_produk, harga, jumlah_terjual, total_harga, total_items, total_amount, transaction_date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		transaksiID,          // Foreign key from transaksi_header
		namaProduk,           // Product name
		harga,                // Product price
		quantitySold,         // Quantity sold
		totalHarga,           // Total price for this product
		quantitySold,         // Total items (same as quantitySold in this case)
		totalHarga,           // Total amount (same as totalHarga in this case)
		time.Now(),           // Transaction date
	)
	if err != nil {
		fmt.Printf("Error inserting into transaksi_detail table: %v\n", err)
		tx.Rollback()
		return fmt.Errorf("failed to insert into transaksi_detail table: %w", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func Api_addSoldItems(w http.ResponseWriter, r *http.Request) {
    db := Koneksi()
    defer db.Close()

    var soldItems []SoldItem
    err := json.NewDecoder(r.Body).Decode(&soldItems)
    if err != nil {
        fmt.Println("Error decoding JSON:", err) // Add this line
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "Error decoding JSON: " + err.Error()}) // Include the error message
        return
    }

    // Validate that we have items to process
    if len(soldItems) == 0 {
        fmt.Println("No items to process") // Add this line
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "No items to process"})
        return
    }

    // Start a transaction for creating the header
    tx, err := db.Begin()
    if err != nil {
        fmt.Println("Failed to start transaction:", err) // Add this line
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "Failed to start transaction: " + err.Error()})
        return
    }

    // Calculate total items and total amount for all products
    totalItems := 0
    var totalAmounts []float64
    var totalAmount float64

    // First pass: collect information about all products
    for _, item := range soldItems {
        // Get product details to calculate total price
        var namaProduk string
        var harga float64
        fmt.Println("Item:", item) // Add this line
        err := db.QueryRow("SELECT nama, harga FROM produk WHERE produk_id = $1", item.ProdukID).Scan(&namaProduk, &harga)
        if err != nil {
            fmt.Println("Failed to get product details:", err) // Add this line
            tx.Rollback()
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get product details: " + err.Error()})
            return
        }

        itemTotal := harga * float64(item.StokKeluar)
        totalItems += item.StokKeluar
        totalAmounts = append(totalAmounts, itemTotal)
        totalAmount += itemTotal
    }

    // Create a single transaksi_header entry for all items
    var transaksiID int
    err = tx.QueryRow(
        "INSERT INTO transaksi_header (total_item, total_jumlah, tanggal_transaksi) VALUES ($1, $2, $3) RETURNING transaksi_id",
        totalItems,
        pq.Array(totalAmounts), // All individual item totals
        time.Now(),
    ).Scan(&transaksiID)
    if err != nil {
        fmt.Println("Failed to create transaction header:", err) // Add this line
        tx.Rollback()
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create transaction header: " + err.Error()})
        return
    }

    // Commit the transaction header
    err = tx.Commit()
    if err != nil {
        fmt.Println("Failed to commit transaction header:", err) // Add this line
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "Failed to commit transaction header: " + err.Error()})
        return
    }

    // Now process each item with the same transaction ID
    for _, item := range soldItems {
        err := updateStockAndRecordSale(item.ProdukID, item.StokKeluar, transaksiID)
        if err != nil {
            fmt.Println("Failed to update stock and record sale:", err) // Add this line
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update stock and record sale: " + err.Error()})
            return
        }
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Sold items processed successfully",
        "transaction_id": fmt.Sprintf("%d", transaksiID),
    })
}

// @Summary Endpoint to handle sold products
// @Description Decreases stock in 'produk' table and records sale in 'stok_keluar' table
// @Tags Produk
// @Accept json
// @Produce json
// @Param soldItems body []SoldItem true "List of sold items"
// @Success 200 {string} string "Sold items processed successfully"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /soldproduk [post]
func Api_soldProduk(w http.ResponseWriter, r *http.Request) {
	Api_addSoldItems(w, r)
}

// @Summary Ambil produk berdasarkan ID
// @Description Mengambil data produk berdasarkan ID
// @Tags Produk
// @Accept json
// @Produce json
// @Param id query string true "Product ID"
// @Success 200 {object} Produk
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /selectprodukByid [get]
func Api_selectProdukById(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	produk, err := getProdukById(id)
	if err != nil {
		// If product not found, return an empty array
		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(nil) // Return null
			return
		}
		http.Error(w, "Failed to get produk: "+err.Error(), http.StatusInternalServerError)
		return
	}

	dataJson, err := json.Marshal(produk) // Return single object
	if err != nil {
		http.Error(w, "Failed to marshal produk: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(dataJson)
}

// @Summary Ambil semua produk
// @Description Mengambil semua data produk dari database
// @Security BearerAuth
// @Tags Produk
// @Accept json
// @Produce json
// @Success 200 {array} Produk
// @Router /selectproduk [get]
func Api_selectAllProduk(w http.ResponseWriter, r *http.Request) {
	produk := select_allProduk()

	dataJson, err := json.Marshal(produk)

	if err != nil {
		panic(err.Error())
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(dataJson)

	}
	for _, produk := range produk {
		fmt.Println(produk)
	}
}

// @Summary Tambah produk baru
// @Description Menambahkan data produk ke dalam database
// @Tags Produk
// @Accept json
// @Produce json
// @Param produk body Produk true "Data produk"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /addproduk [post]
// @Security BearerAuth
func API_addProduk(w http.ResponseWriter, r *http.Request) {
	var produk Produk
	err := json.NewDecoder(r.Body).Decode(&produk)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	addDataProduk(produk.Nama, produk.Stok, produk.Harga, produk.HargaBeli, produk.Foto, produk.Supplier)
}

// @Summary Update data produk
// @Description Memperbarui informasi produk berdasarkan ID
// @Tags Produk
// @Accept json
// @Produce json
// @Param produk body Produk true "Data produk yang diperbarui"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /updateproduk [put]
// @Security BearerAuth
func Api_updateProduk(w http.ResponseWriter, r *http.Request) {
	var produk Produk
	err := json.NewDecoder(r.Body).Decode(&produk)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updateDataProduk(produk.ProdukID, produk.Nama, produk.Stok, produk.Harga, produk.HargaBeli, produk.Foto, produk.Supplier)
}

// @Summary Hapus produk
// @Description Menghapus produk berdasarkan ID
// @Tags Produk
// @Accept json
// @Produce json
// @Param produk body Produk true "Produk yang akan dihapus (berdasarkan ID)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /deleteproduk [delete]
// @Security BearerAuth
func Api_deleteProduk(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	var produk Produk
	err := json.NewDecoder(r.Body).Decode(&produk)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	deleteDataProduk(fmt.Sprint(produk.ProdukID))
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	db := Koneksi()
	defer db.Close()

	// SQL query to fetch transaction history
	sqlQuery := `
		SELECT 
            td.transaksi_id, 
            td.nama_produk, 
            td.harga AS harga_jual, 
            p.harga_beli, 
            td.jumlah_terjual, 
            td.total_harga, 
            th.tanggal_transaksi
        FROM transaksi_detail td
        INNER JOIN produk p ON td.nama_produk = p.nama
        INNER JOIN transaksi_header th ON td.transaksi_id = th.transaksi_id
        ORDER BY td.transaksi_id DESC
    `
	rows, err := db.Query(sqlQuery)
	if err != nil {
		http.Error(w, "Failed to fetch transaction history: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var transactionID string
		var namaProduk string
		var hargaJual float64
		var hargaBeli float64
		var jumlahTerjual int
		var totalHarga float64
		var tanggalTransaksi time.Time

		err := rows.Scan(&transactionID, &namaProduk, &hargaJual, &hargaBeli, &jumlahTerjual, &totalHarga, &tanggalTransaksi)
		if err != nil {
			http.Error(w, "Failed to scan transaction row: "+err.Error(), http.StatusInternalServerError)
			return
		}

		transaction := map[string]interface{}{
			"transaksi_id":      transactionID,
			"nama_produk":       namaProduk,
			"harga_jual":        hargaJual,
			"harga_beli":        hargaBeli,
			"jumlah_terjual":    jumlahTerjual,
			"total_harga":       totalHarga,
			"tanggal_transaksi": tanggalTransaksi,
		}
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating through rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

func GetDailySales(w http.ResponseWriter, r *http.Request) {
    db := Koneksi()
    defer db.Close()

    // Get the date from the query parameters
    dateStr := r.URL.Query().Get("date")
    if dateStr == "" {
        // If no date is provided, use the current date
        dateStr = time.Now().Format("2006-01-02")
    }

    // Parse the date string
    date, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        http.Error(w, "Invalid date format. Use YYYY-MM-DD.", http.StatusBadRequest)
        return
    }

    // SQL query to calculate total sales for the given date
    sqlQuery := `
        SELECT SUM(total_harga)
        FROM transaksi
        WHERE DATE(tanggal) = $1
    `

    var totalSales float64
    err = db.QueryRow(sqlQuery, date).Scan(&totalSales)
    if err != nil {
        if err == sql.ErrNoRows {
            // If there are no sales for the given date, return 0
            totalSales = 0
        } else {
            http.Error(w, "Failed to fetch daily sales: "+err.Error(), http.StatusInternalServerError)
            return
        }
    }

    // Create a map to hold the result
    result := map[string]interface{}{
        "date":       date.Format("2006-01-02"),
        "totalSales": totalSales,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

func GetWeeklySales(w http.ResponseWriter, r *http.Request) {
    db := Koneksi()
    defer db.Close()

    // Get the start and end dates from the query parameters
    startDateStr := r.URL.Query().Get("startDate")
    endDateStr := r.URL.Query().Get("endDate")

    // Parse the date strings
    startDate, err := time.Parse("2006-01-02", startDateStr)
    if err != nil {
        http.Error(w, "Invalid start date format. Use YYYY-MM-DD.", http.StatusBadRequest)
        return
    }

    endDate, err := time.Parse("2006-01-02", endDateStr)
    if err != nil {
        http.Error(w, "Invalid end date format. Use YYYY-MM-DD.", http.StatusBadRequest)
        return
    }

    // SQL query to calculate total sales for each day in the given date range
    sqlQuery := `
        SELECT DATE(tanggal), SUM(total_harga)
        FROM transaksi
        WHERE DATE(tanggal) >= $1 AND DATE(tanggal) <= $2
        GROUP BY DATE(tanggal)
        ORDER BY DATE(tanggal)
    `

    rows, err := db.Query(sqlQuery, startDate, endDate)
    if err != nil {
        http.Error(w, "Failed to fetch weekly sales: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var sales []map[string]interface{}
    for rows.Next() {
        var date time.Time
        var salesAmount float64

        err := rows.Scan(&date, &salesAmount)
        if err != nil {
            http.Error(w, "Failed to scan sales row: "+err.Error(), http.StatusInternalServerError)
            return
        }

        sale := map[string]interface{}{
            "date":  date.Format("2006-01-02"),
            "sales": salesAmount,
        }
        sales = append(sales, sale)
    }

    if err := rows.Err(); err != nil {
        http.Error(w, "Error iterating through rows: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(sales)
}

func GetMonthlyRevenue(w http.ResponseWriter, r *http.Request) {
    db := Koneksi()
    defer db.Close()

    // Get the current year and month
    now := time.Now()
    year, month, _ := now.Date()

    // Calculate the start and end dates of the month
    firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
    lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

    // SQL query to calculate total revenue for the current month
    sqlQuery := `
        SELECT SUM(total_harga)
        FROM transaksi
        WHERE tanggal >= $1 AND tanggal <= $2
    `

    var totalRevenue float64
    err := db.QueryRow(sqlQuery, firstOfMonth, lastOfMonth).Scan(&totalRevenue)
    if err != nil {
        if err == sql.ErrNoRows {
            // If there are no sales for the month, return 0
            totalRevenue = 0
        } else {
            http.Error(w, "Failed to fetch monthly revenue: "+err.Error(), http.StatusInternalServerError)
            return
        }
    }

    // Create a map to hold the result
    result := map[string]interface{}{
        "month":       month.String(),
        "year":        year,
        "totalRevenue": totalRevenue,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

func GetInventoryStatus(w http.ResponseWriter, r *http.Request) {
    db := Koneksi()
    defer db.Close()

    // SQL query to count the total number of products
    sqlQuery := `
        SELECT COUNT(*)
        FROM produk
    `

    var totalProducts int
    err := db.QueryRow(sqlQuery).Scan(&totalProducts)
    if err != nil {
        http.Error(w, "Failed to fetch inventory status: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Create a map to hold the result
    result := map[string]interface{}{
        "totalProducts": totalProducts,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

