package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/shaiksadikjanu-cmd/mini-render/db"
	"github.com/shaiksadikjanu-cmd/mini-render/handlers"
)

func loadEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	lines := splitLines(string(data))
	for _, line := range lines {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		for i, ch := range line {
			if ch == '=' {
				os.Setenv(line[:i], line[i+1:])
				break
			}
		}
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func main() {
	loadEnv()
	db.Init()

	fmt.Println("♻️  restoring previous deployments...")
	handlers.RestoreDeployments()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.Handle("/", http.FileServer(http.Dir("frontend/static")))
	http.HandleFunc("/api/signup", cors(handlers.Signup))
	http.HandleFunc("/api/login", cors(handlers.Login))
	http.HandleFunc("/api/deploy", cors(handlers.Deploy))
	http.HandleFunc("/api/deploy-code", cors(handlers.DeployCode))
	http.HandleFunc("/api/status", cors(handlers.Status))

	fmt.Printf("🚀 mini-render running on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
