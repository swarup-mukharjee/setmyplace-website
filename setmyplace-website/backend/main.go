// Set My Place — contact form backend (PostgreSQL edition)
//
// One external dependency: github.com/lib/pq (the Postgres driver).
// Everything else is standard library.
//
// First-time setup:
//
//	cd backend
//	go mod tidy      // fetches lib/pq and writes go.sum
//	go run main.go
//
// Build a binary:
//
//	go build -o setmyplace-server main.go
//	./setmyplace-server
//
// Environment variables:
//
//	DATABASE_URL  - Postgres connection string, e.g.
//	                postgres://user:password@localhost:5432/setmyplace?sslmode=disable
//	PORT          - port to listen on (default 8080)
//	ADMIN_KEY     - secret key required to read submissions via /api/contacts
//	SITE_DIR      - folder to serve the website from (default "../frontend")
//
// Endpoints:
//
//	GET  /                serves the static website (index.html etc.)
//	POST /api/contact     accepts a contact form submission (JSON)
//	GET  /api/contacts    lists all submissions (requires X-Admin-Key header
//	                      or ?key= query param matching ADMIN_KEY)
//	GET  /api/health      simple health check, confirms DB connectivity
package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// ContactSubmission is one row in the contacts table.
type ContactSubmission struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Service   string `json:"service"`
	Date      string `json:"date"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
	IP        string `json:"ip"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updated_at"`
}

type contactStatusPayload struct {
	Status string `json:"status"`
}

// incoming payload from the contact form
type contactPayload struct {
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	Service        string `json:"service"`
	Date           string `json:"date"`
	Message        string `json:"message"`
	CompanyWebsite string `json:"company_website"` // honeypot field, must stay empty
}

var db *sql.DB

func main() {

	// Load local environment variables from .env.
	// If .env doesn't exist, continue using system environment variables.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using system environment variables.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	adminKey := os.Getenv("ADMIN_KEY")
	if adminKey == "" {
		adminKey = "changeme"
		log.Println("WARNING: ADMIN_KEY is not set. Using the default 'changeme'.")
		log.Println("         Set a real ADMIN_KEY before deploying this anywhere public.")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set. Example:\n" +
			`  export DATABASE_URL="postgres://user:password@localhost:5432/setmyplace?sslmode=disable"`)
	}

	siteDir := os.Getenv("SITE_DIR")
	if siteDir == "" {
		siteDir = "../frontend" // frontend folder, relative to /backend
	}

	assetsDir := os.Getenv("ASSETS_DIR")
	if assetsDir == "" {
		assetsDir = "../assets" // assets folder, relative to /backend
	}

	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("could not open database connection: %v", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("could not connect to Postgres: %v\nCheck DATABASE_URL and that Postgres is running.", err)
	}
	log.Println("Connected to Postgres.")

	if err := ensureSchema(); err != nil {
		log.Fatalf("could not set up database schema: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", withCORS(handleHealth))
	mux.HandleFunc("/api/contact", withCORS(handleContact))
	mux.HandleFunc("/api/contacts", withCORS(handleListContacts(adminKey)))
	mux.HandleFunc("/api/contacts/", withCORS(handleUpdateContact(adminKey)))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir))))
	mux.Handle("/", http.FileServer(http.Dir(siteDir)))

	addr := ":" + port
	log.Printf("Set My Place backend running at http://localhost%s", addr)
	log.Printf("Serving website files from: %s", siteDir)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// ensureSchema creates the contacts table if it doesn't already exist,
// so there's no separate migration step to run by hand.
func ensureSchema() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS contacts (
			id             BIGSERIAL PRIMARY KEY,
			name           TEXT NOT NULL,
			phone          TEXT NOT NULL,
			service        TEXT,
			preferred_date TEXT,
			message        TEXT,
			created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
			ip             TEXT,
			status         TEXT NOT NULL DEFAULT 'new',
			updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		ALTER TABLE contacts ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'new';
		ALTER TABLE contacts ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
		CREATE INDEX IF NOT EXISTS idx_contacts_created_at ON contacts (created_at DESC);
	`)
	return err
}

// withCORS allows the API to be called from a different origin too,
// in case the frontend is ever hosted separately from this backend.
func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Admin-Key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := db.Ping(); err != nil {
		respondJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db unreachable"})
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var payload contactPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Honeypot: real visitors never fill this hidden field. If it has a
	// value, silently pretend success so bots don't learn they were caught.
	if strings.TrimSpace(payload.CompanyWebsite) != "" {
		respondJSON(w, http.StatusOK, map[string]bool{"success": true})
		return
	}

	payload.Name = strings.TrimSpace(payload.Name)
	payload.Phone = strings.TrimSpace(payload.Phone)
	if payload.Name == "" || payload.Phone == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "name and phone are required"})
		return
	}
	if len(payload.Name) > 200 || len(payload.Phone) > 40 || len(payload.Message) > 4000 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "one or more fields are too long"})
		return
	}

	ip := clientIP(r)

	var id int64
	var createdAt time.Time
	err := db.QueryRow(
		`INSERT INTO contacts (name, phone, service, preferred_date, message, ip)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		payload.Name,
		payload.Phone,
		strings.TrimSpace(payload.Service),
		strings.TrimSpace(payload.Date),
		strings.TrimSpace(payload.Message),
		ip,
	).Scan(&id, &createdAt)

	if err != nil {
		log.Println("insert error:", err)
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not save submission, please try again"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func handleListContacts(adminKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		key := r.Header.Get("X-Admin-Key")
		if key == "" {
			key = r.URL.Query().Get("key")
		}
		if key != adminKey {
			respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		rows, err := db.Query(`
			SELECT id, name, phone, COALESCE(service,''), COALESCE(preferred_date,''),
			       COALESCE(message,''), created_at, COALESCE(ip,''), COALESCE(status,'new'), updated_at
			FROM contacts
			ORDER BY created_at DESC
		`)
		if err != nil {
			log.Println("query error:", err)
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not read submissions"})
			return
		}
		defer rows.Close()

		subs := []ContactSubmission{}
		for rows.Next() {
			var s ContactSubmission
			var createdAt, updatedAt time.Time
			if err := rows.Scan(&s.ID, &s.Name, &s.Phone, &s.Service, &s.Date, &s.Message, &createdAt, &s.IP, &s.Status, &updatedAt); err != nil {
				log.Println("scan error:", err)
				respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not read submissions"})
				return
			}
			s.CreatedAt = createdAt.UTC().Format(time.RFC3339)
			s.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
			subs = append(subs, s)
		}

		respondJSON(w, http.StatusOK, subs)
	}
}

func handleUpdateContact(adminKey string) http.HandlerFunc {
	validStatuses := map[string]bool{"new": true, "contacted": true, "accepted": true, "attended": true}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		if r.Header.Get("X-Admin-Key") != adminKey {
			respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/contacts/"), 10, 64)
		if err != nil || id < 1 {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid contact id"})
			return
		}

		var payload contactStatusPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || !validStatuses[payload.Status] {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid status"})
			return
		}

		result, err := db.Exec(`UPDATE contacts SET status = $1, updated_at = now() WHERE id = $2`, payload.Status, id)
		if err != nil {
			log.Println("update error:", err)
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not update submission"})
			return
		}
		count, _ := result.RowsAffected()
		if count == 0 {
			respondJSON(w, http.StatusNotFound, map[string]string{"error": "submission not found"})
			return
		}
		respondJSON(w, http.StatusOK, map[string]bool{"success": true})
	}
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.Split(fwd, ",")[0]
	}
	return r.RemoteAddr
}

func respondJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
