package database

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var secretKey = []byte("rahasia")
var api_key = "123"


func generateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour).Unix()
	tokenStr, err := token.SignedString(secretKey)
	if err != nil {
		panic(err.Error())
	} 
	return tokenStr, err

}

func MiddleWare(next func(w http.ResponseWriter, r *http.Request))http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["token"] != nil {
			token, err := jwt.Parse(r.Header["token"][0], func(t *jwt.Token) (interface{}, error) {
				_, hasil := t.Method.(*jwt.SigningMethodHMAC) 
				if !hasil {
					// tampilkan pesan error apabila hasil gagal
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("gagal autentikasi"))
				}
				return secretKey, nil
			})
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("gagal autentikasi"))
				panic(err.Error())
			}
			if token.Valid {
				next(w, r)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("gagal autentikasi"))
			}
		}
	})
}

func API_generateJWT(h http.ResponseWriter, r *http.Request) {
	token, err := generateJWT()
	tokenStr, err := json.Marshal(token)
	if err != nil {
		panic(err.Error())
	} else {
		h.Header().Set("Content-Type", "application/json")
		io.WriteString(h, string(tokenStr))
	}

}

	
