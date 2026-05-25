package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/shaiksadikjanu-cmd/mini-render/db"
)

type User struct {
	ID        string    `json:"id,omitempty"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	Plan      string    `json:"plan,omitempty"`
}

func hashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password + os.Getenv("JWT_SECRET")))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respond(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if input.Name == "" || input.Email == "" || input.Username == "" || input.Password == "" {
		respond(w, http.StatusBadRequest, map[string]string{"error": "all fields are required"})
		return
	}

	newUser := User{
		Name:     input.Name,
		Email:    input.Email,
		Username: input.Username,
		Password: hashPassword(input.Password),
		Plan:     "free",
	}

	data, err := db.Insert("users", newUser)
	if err != nil {
		respond(w, http.StatusConflict, map[string]string{"error": "email or username already exists"})
		return
	}

	var created []User
	json.Unmarshal(data, &created)

	if len(created) > 0 {
		created[0].Password = ""
		respond(w, http.StatusCreated, map[string]interface{}{
			"message": "account created successfully",
			"user":    created[0],
		})
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respond(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	filter := fmt.Sprintf("email=eq.%s&password=eq.%s", input.Email, hashPassword(input.Password))
	data, err := db.Select("users", filter)
	if err != nil {
		respond(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	var users []User
	json.Unmarshal(data, &users)

	if len(users) == 0 {
		respond(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}

	user := users[0]
	user.Password = ""

	respond(w, http.StatusOK, map[string]interface{}{
		"message": "login successful",
		"user":    user,
	})
}
