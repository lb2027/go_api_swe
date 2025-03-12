package database

import (
	_ "github.com/lib/pq"
)

type Produk struct {
	ProdukID  string `json:"produk_id"`
	Nama      string `json:"name"`
	Stok      int    `json:"stok"`
	Harga     int    `json:"harga"`
	HargaBeli int    `json:"harga_beli"`
	Foto      string `json:"foto"`
	Supplier  string `json:"supplier"`
}

func select_allProduk() []Produk {
	db:=Koneksi()
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

func getProdukById(id string) []Produk {
	db := Koneksi()
	var produk = Produk{}
	var produkArray []Produk

	statement, err := db.Query("SELECT produk_id, nama, stok, harga, harga_beli, foto, supplier FROM produk WHERE produk_id = $1", id)

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

func getProdukByname(name string) []Produk {
	db := Koneksi()
	var produk = Produk{}
	var produkArray []Produk

	statement, err := db.Query("SELECT produk_id, nama, stok, harga, harga_beli, foto, supplier FROM produk WHERE nama = $1", name)

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

func addDataProduk(nama string, stok int, harga int, harga_beli int, foto string, supplier string) {
	db := Koneksi()
	statement, err := db.Prepare("INSERT INTO produk (nama, stok, harga, harga_beli, foto, supplier) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		panic(err.Error())
	} else {
		statement.Exec(nama, stok, harga, harga_beli, foto, supplier)
	}
}



