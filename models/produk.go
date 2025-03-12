package models

type Produk struct {
	ProdukID string `json:"produk_id"`
	Nama string `json:"name"`
	Stok int `json:"stok"`
	Harga int `json:"harga"`
	HargaBeli int `json:"harga_beli"`
	Foto string `json:"foto"`
	Supplier string `json:"supplier"`
}

