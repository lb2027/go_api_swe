package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
)

type User struct {
	User_id      int    `json:"id"`
	Username    string `json:"username"`
	Password   	string `json:"password"`
	Role 		string `json:"role"`
}
// Api_selectAllData godoc
// @Summary Get all user
// @Description Get list of all user from database
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {array} User
// @Router /user [get]
func Select_alluser() []User {
	db := Koneksi()
	var user = User{}
	var userArray []User

	statement, err := db.Query("SELECT * FROM users")

	if err != nil {
		panic(err.Error())
	} else {
		for statement.Next() {
			err = statement.Scan(&user.User_id, &user.Username, &user.Password, &user.Role)
			if err != nil {
				panic(err.Error())
			} else {
				userArray = append(userArray, User{User_id: user.User_id, Username: user.Username, Password: user.Password, Role: user.Role})
			}
		}
	}
	return userArray
}



func addData(username string, password string, role string) {
    db := Koneksi()
    sttmnt, err := db.Prepare("INSERT INTO users(username, password, role) VALUES($1, $2, $3)")
    if err != nil {
        panic(err.Error())
    }
    _, err = sttmnt.Exec(username, password, role)
    if err != nil {
        panic(err.Error())
    }
}


// connect
func Koneksi() *sql.DB {
	const (
		host     = "localhost"
		port     = 5432
		user     = "postgres"
		password = "12460"
		dbname   = "frozen_food"
	)
	
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlconn)

    if err != nil {
        panic(err.Error())
    } else {
        fmt.Println("connected")
        return db
    }
}

// @Summary Menambahkan user baru
// @Description Menambahkan data user ke database
// @Tags User
// @Accept json
// @Produce json
// @Param user body User true "Data user"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /adduser [post]
func API_add(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	addData(user.Username, user.Password, user.Role)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "User added successfully"}`))
}

func Api_selectAllData(w http.ResponseWriter, r *http.Request) {

	dataUser := Select_alluser()
	dataJson, err := json.Marshal(dataUser)
	if err != nil {
		panic(err.Error())
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(dataJson)

	}

}

// @Summary Hapus user
// @Description Menghapus user berdasarkan ID dari body JSON
// @Tags User
// @Accept json
// @Produce json
// @Param user body User true "ID user yang akan dihapus"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} map[string]string
// @Router /deleteuser [delete]
func Api_deleteUser(w http.ResponseWriter, r *http.Request) {
    var user User
    err := json.NewDecoder(r.Body).Decode(&user)

    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if user.User_id == 0 {
        http.Error(w, "no id found", http.StatusBadRequest)
        return
    }

    db := Koneksi()
    statement, err := db.Prepare("DELETE FROM users WHERE user_id = $1")
    if err != nil {
        panic(err.Error())
    } else {
        _, err = statement.Exec(user.User_id)
        if err != nil {
            panic(err.Error())
        }
    }
    w.WriteHeader(http.StatusNoContent)
}

// @Summary Update user
// @Description Memperbarui data user berdasarkan ID
// @Tags User
// @Accept json
// @Produce json
// @Param user body User true "Data user yang diperbarui"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} map[string]string
// @Router /updateuser [put]
func Api_updateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user.User_id == 0 {
		http.Error(w, "no id found", http.StatusBadRequest)
		return
	}

	db := Koneksi()
	statement, err := db.Prepare("UPDATE users SET username = $1, password = $2, role = $3 WHERE user_id = $4")
	if err != nil {
		panic(err.Error())
	} else {
		_, err = statement.Exec(user.Username, user.Password, user.Role, user.User_id)
		if err != nil {
			panic(err.Error())
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// func Api_getNameAddress(w http.ResponseWriter, r *http.Request) {
// 	address := r.URL.Query().Get("address")
// 	if address == "" {	
// 		http.Error(w, "no id found", http.StatusBadRequest)
// 		return
// 	}

// 	user := getNamenAdress(address)
// 	dataJson, err := json.Marshal(user)
// 	if err != nil {
// 		panic(err.Error())
// 	} else {
// 		io.WriteString(w, string(dataJson))
// 	}

// }