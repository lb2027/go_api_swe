package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

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
	ProdukID   int `json:"produk_id"`
	StokKeluar int `json:"stok_keluar"`
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
	statement, err := db.Prepare("UPDATE produk SET nama = $1, stok = $2, harga = $3, harga_beli = $4, foto = $5, supplier = $6 WHERE produk_id = $7")
	if err != nil {
		panic(err.Error())
	} else {
		statement.Exec(nama, stok, harga, harga_beli, foto, supplier, id)
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

func updateStock(productID int, quantitySold int) error {
	db := Koneksi()
	defer db.Close()

	_, err := db.Exec("UPDATE produk SET stok = stok - $1 WHERE produk_id = $2", quantitySold, productID)
	return err
}

func Api_addSoldItems(w http.ResponseWriter, r *http.Request) {
	var soldItems []SoldItem
	err := json.NewDecoder(r.Body).Decode(&soldItems)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, item := range soldItems {
		err := updateStock(item.ProdukID, item.StokKeluar)
		if err != nil {
			http.Error(w, "Failed to update stock: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Sold items added successfully")
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


