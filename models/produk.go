package models

import "gorm.io/gorm"

type Produk struct {
	gorm.Model
	produk_id string `json:"produk_id"`
	nama string `json:"name"`
	stok int `json:"stok"`
	harga int `json:"harga"`
	harga_beli int `json: "harga_beli"`
	foto string `json:"foto"`
	supplier string `json:"supplier"`
}

