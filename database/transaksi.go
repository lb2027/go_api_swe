package database

import (
	"encoding/json"
	"net/http"
	"time"
)

type Transaksi struct {
	TransaksiID   int     `json:"transaksi_id"`
	NamaProduk    string  `json:"nama_produk"`
	Harga         float64 `json:"harga"`
	JumlahTerjual int     `json:"jumlah_terjual"`
	TotalHarga    float64 `json:"total_harga"`
	Tanggal      time.Time`json:"tanggal"`
}

func Select_allTransaksi() []Transaksi {
	db := Koneksi()
	rows, err := db.Query("SELECT * FROM transaksi")
	if err != nil {
		panic(err.Error())
	}
	var dataTransaksi []Transaksi
	for rows.Next() {
		var transaksi Transaksi
		err = rows.Scan(&transaksi.TransaksiID, &transaksi.NamaProduk, &transaksi.Harga, &transaksi.JumlahTerjual, &transaksi.TotalHarga, &transaksi.Tanggal)
		if err != nil {
			panic(err.Error())
		}
		dataTransaksi = append(dataTransaksi, transaksi)
	}
	return dataTransaksi
}

// @Summary Ambil semua transaksi
// @Description Mengambil semua data transaksi dari database
// @Tags Transaksi
// @Accept json
// @Produce json
// @Success 200 {array} Transaksi
// @Router /selecttransaksi [get]
// @Security BearerAuth
func Api_selectAllTransaksi(w http.ResponseWriter, r *http.Request) {
	dataTransaksi := Select_allTransaksi()
	dataJson, err := json.Marshal(dataTransaksi)
	if err != nil {
		panic(err.Error())
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(dataJson)
	}
}

func AddDataTransaksi(namaProduk string, harga float64, jumlahTerjual int, totalHarga float64, tanggal string) {
    db := Koneksi()
    parsedDate, err := time.Parse(time.RFC3339, tanggal) // Parse the date string
    if err != nil {
        panic(err.Error())
    }
    statement, err := db.Prepare("INSERT INTO transaksi (nama_produk, harga, jumlah_terjual, total_harga, tanggal) VALUES ($1, $2, $3, $4, $5)")
    if err != nil {
        panic(err.Error())
    } else {
        statement.Exec(namaProduk, harga, jumlahTerjual, totalHarga, parsedDate)
    }
}

// @Summary Tambah transaksi
// @Description Menambahkan data transaksi baru
// @Tags Transaksi
// @Accept json
// @Produce json
// @Param transaksi body Transaksi true "Data transaksi"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /addtransaksi [post]
// @Security BearerAuth
func Api_addTransaksi(w http.ResponseWriter, r *http.Request) {
    var transaksi Transaksi
    err := json.NewDecoder(r.Body).Decode(&transaksi)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
	AddDataTransaksi(transaksi.NamaProduk, transaksi.Harga, transaksi.JumlahTerjual, transaksi.TotalHarga, transaksi.Tanggal.Format(time.RFC3339))
}