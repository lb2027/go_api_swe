package models

import "time"

type Transaksi struct {
	transaksi_id   int     `json:"transaksi_id"`
	nama_produk    string  `json:"nama_produk"`
	harga          float64 `json:"harga"`
	jumlah_terjual int     `json:"jumlah_terjual"`
	total_harga    float64 `json:"total_harga"`
	tanggal        time.Time `json:"tanggal"`
}

