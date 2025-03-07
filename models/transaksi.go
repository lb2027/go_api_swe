package models

import "time"

type Transaksi struct {
	TransaksiID   int     `json:"transaksi_id"`
	NamaProduk    string  `json:"nama_produk"`
	Harga          float64 `json:"harga"`
	JumlahTerjual int     `json:"jumlah_terjual"`
	TotalHarga    float64 `json:"total_harga"`
	Tanggal        time.Time `json:"tanggal"`
}
