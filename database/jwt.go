package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var secretKey = []byte("rahasia")
var api_key = "123"

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func generateJWT(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username // Store username in token
	claims["exp"] = time.Now().Add(time.Hour).Unix()
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// @Summary Generate JWT Token
// @Description Menghasilkan JWT token untuk autentikasi API
// @Tags Auth
// @Accept json
// @Produce json
// @Param creds body Credentials true "User credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /login [post]
func API_generateJWT(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// **Authentication Logic**
	isAuthenticated, err := AuthenticateUser(creds.Username, creds.Password)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	if !isAuthenticated {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// **Retrieve the user's role from the database**
	role, err := GetUserRole(creds.Username) // Implement this function
	if err != nil {
		http.Error(w, "Failed to get user role", http.StatusInternalServerError)
		return
	}

	tokenString, err := generateJWT(creds.Username)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// **Include the role in the response**
	response := map[string]string{
		"token": tokenString,
		"role":  role, // Add the role to the response
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Implement this function to retrieve the user's role from the database
func GetUserRole(username string) (string, error) {
	db := Koneksi()
	defer db.Close()

	var role string
	err := db.QueryRow("SELECT role FROM users WHERE username = $1", username).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("user not found")
		}
		return "", err
	}

	return role, nil
}

func AuthenticateUser(username, password string) (bool, error) {
	// Implement your authentication logic here
	// Example:
	db := Koneksi()
	defer db.Close()

	var storedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // User not found
		}
		return false, err // Database error
	}

	// Compare the provided password with the stored password
	// You should use a secure password hashing algorithm (e.g., bcrypt)
	if password == storedPassword {
		return true, nil
	}

	return false, nil
}

func RegisterUser(username, password, role string) error {
	db := Koneksi()
	defer db.Close()

	// Hash the password before storing it
	// You should use a secure password hashing algorithm (e.g., bcrypt)
	// For simplicity, this example stores the password in plain text
	_, err := db.Exec("INSERT INTO users (username, password, role) VALUES ($1, $2, $3)", username, password, role)
	if err != nil {
		return err
	}

	return nil
}

// @Summary Register a new user
// @Description Registers a new user in the database
// @Tags Auth
// @Accept json
// @Produce json
// @Param creds body Credentials true "User credentials"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /register [post]
func API_register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set a default role for new users
	role := "user"

	err = RegisterUser(creds.Username, creds.Password, role)
	if err != nil {
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "User registered successfully"}`))
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


