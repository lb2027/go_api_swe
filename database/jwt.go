package database

import (
	"encoding/json"
	"fmt"
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

func MiddleWare(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Print headers for debugging
		fmt.Println("Received headers:", r.Header)

		tokenHeader := r.Header.Get("token")
		if tokenHeader == "" {
			http.Error(w, "No token found in header", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenHeader, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return secretKey, nil
		})

		if err != nil {
			fmt.Println("Token parsing error:", err)
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		if token.Valid {
			fmt.Println("Token is valid, proceeding with request")
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		}
	})
}

// @Summary Generate JWT Token
// @Description Menghasilkan JWT token untuk autentikasi API
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 405 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /login [post]
func API_generateJWT(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token, err := generateJWT()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"token": token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


