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
//	POST /api/partner-registration accepts a partner/vendor registration (JSON)
//	GET  /api/contacts    lists all submissions (requires X-Admin-Key header
//	                      or ?key= query param matching ADMIN_KEY)
//	GET  /api/partner-registrations lists partner/vendor registrations (same auth)
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
	"github.com/lib/pq"
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

type PartnerRegistration struct {
	ID                  int64    `json:"id"`
	FullName            string   `json:"full_name"`
	GuardianName        string   `json:"guardian_name"`
	DateOfBirth         string   `json:"date_of_birth"`
	MobileNumber        string   `json:"mobile_number"`
	AlternateNumber     string   `json:"alternate_number"`
	Email               string   `json:"email"`
	CurrentAddress      string   `json:"current_address"`
	PermanentAddress    string   `json:"permanent_address"`
	City                string   `json:"city"`
	State               string   `json:"state"`
	PIN                 string   `json:"pin"`
	EmergencyContact    string   `json:"emergency_contact"`
	PartnerTypes        []string `json:"partner_types"`
	ServiceCategories   []string `json:"service_categories"`
	PrimaryService      string   `json:"primary_service"`
	YearsExperience     string   `json:"years_experience"`
	PreferredLocations  string   `json:"preferred_locations"`
	Languages           string   `json:"languages"`
	Skills              string   `json:"skills"`
	DrivingLicence      string   `json:"driving_licence"`
	LicenceExpiry       string   `json:"licence_expiry"`
	VehicleOwner        string   `json:"vehicle_owner"`
	VehicleRegistration string   `json:"vehicle_registration"`
	VehicleType         string   `json:"vehicle_type"`
	VehicleMakeModel    string   `json:"vehicle_make_model"`
	RCNumber            string   `json:"rc_number"`
	InsuranceDetails    string   `json:"insurance_details"`
	PUCValidUntil       string   `json:"puc_valid_until"`
	PermitDetails       string   `json:"permit_details"`
	BusinessName        string   `json:"business_name"`
	BusinessType        string   `json:"business_type"`
	AuthorizedPerson    string   `json:"authorized_person"`
	BusinessAddress     string   `json:"business_address"`
	GSTIN               string   `json:"gstin"`
	PAN                 string   `json:"pan"`
	TradeRegistration   string   `json:"trade_registration"`
	Employees           string   `json:"employees"`
	TeamSize            string   `json:"team_size"`
	YearsInBusiness     string   `json:"years_in_business"`
	PreviousExperience  string   `json:"previous_experience"`
	GovernmentID        string   `json:"government_id"`
	PANNumber           string   `json:"pan_number"`
	OtherIDProof        string   `json:"other_id_proof"`
	BankAccountHolder   string   `json:"bank_account_holder"`
	BankName            string   `json:"bank_name"`
	AccountNumber       string   `json:"account_number"`
	IFSCCode            string   `json:"ifsc_code"`
	UPIID               string   `json:"upi_id"`
	Documents           []string `json:"documents"`
	Capabilities        []string `json:"capabilities"`
	WorkingDays         string   `json:"working_days"`
	WorkingHours        string   `json:"working_hours"`
	ServiceRadius       string   `json:"service_radius"`
	ReferenceOne        string   `json:"reference_one"`
	ReferenceTwo        string   `json:"reference_two"`
	DeclarationAccepted bool     `json:"declaration_accepted"`
	CreatedAt           string   `json:"created_at"`
	IP                  string   `json:"ip"`
	Status              string   `json:"status"`
}

type contactStatusPayload struct {
	Status string `json:"status"`
}

type partnerRegistrationPayload struct {
	FullName            string   `json:"full_name"`
	GuardianName        string   `json:"guardian_name"`
	DateOfBirth         string   `json:"date_of_birth"`
	MobileNumber        string   `json:"mobile_number"`
	AlternateNumber     string   `json:"alternate_number"`
	Email               string   `json:"email"`
	CurrentAddress      string   `json:"current_address"`
	PermanentAddress    string   `json:"permanent_address"`
	City                string   `json:"city"`
	State               string   `json:"state"`
	PIN                 string   `json:"pin"`
	EmergencyContact    string   `json:"emergency_contact"`
	PartnerTypes        []string `json:"partner_types"`
	ServiceCategories   []string `json:"service_categories"`
	PrimaryService      string   `json:"primary_service"`
	YearsExperience     string   `json:"years_experience"`
	PreferredLocations  string   `json:"preferred_locations"`
	Languages           string   `json:"languages"`
	Skills              string   `json:"skills"`
	DrivingLicence      string   `json:"driving_licence"`
	LicenceExpiry       string   `json:"licence_expiry"`
	VehicleOwner        string   `json:"vehicle_owner"`
	VehicleRegistration string   `json:"vehicle_registration"`
	VehicleType         string   `json:"vehicle_type"`
	VehicleMakeModel    string   `json:"vehicle_make_model"`
	RCNumber            string   `json:"rc_number"`
	InsuranceDetails    string   `json:"insurance_details"`
	PUCValidUntil       string   `json:"puc_valid_until"`
	PermitDetails       string   `json:"permit_details"`
	BusinessName        string   `json:"business_name"`
	BusinessType        string   `json:"business_type"`
	AuthorizedPerson    string   `json:"authorized_person"`
	BusinessAddress     string   `json:"business_address"`
	GSTIN               string   `json:"gstin"`
	PAN                 string   `json:"pan"`
	TradeRegistration   string   `json:"trade_registration"`
	Employees           string   `json:"employees"`
	TeamSize            string   `json:"team_size"`
	YearsInBusiness     string   `json:"years_in_business"`
	PreviousExperience  string   `json:"previous_experience"`
	GovernmentID        string   `json:"government_id"`
	PANNumber           string   `json:"pan_number"`
	OtherIDProof        string   `json:"other_id_proof"`
	BankAccountHolder   string   `json:"bank_account_holder"`
	BankName            string   `json:"bank_name"`
	AccountNumber       string   `json:"account_number"`
	IFSCCode            string   `json:"ifsc_code"`
	UPIID               string   `json:"upi_id"`
	Documents           []string `json:"documents"`
	Capabilities        []string `json:"capabilities"`
	WorkingDays         string   `json:"working_days"`
	WorkingHours        string   `json:"working_hours"`
	ServiceRadius       string   `json:"service_radius"`
	ReferenceOne        string   `json:"reference_one"`
	ReferenceTwo        string   `json:"reference_two"`
	DeclarationAccepted bool     `json:"declaration_accepted"`
	CompanyWebsite      string   `json:"company_website"`
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
	mux.HandleFunc("/api/partner-registration", withCORS(handlePartnerRegistration))
	mux.HandleFunc("/api/partner-registrations", withCORS(handleListPartnerRegistrations(adminKey)))
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
		CREATE TABLE IF NOT EXISTS partner_registrations (
			id BIGSERIAL PRIMARY KEY,
			full_name TEXT NOT NULL,
			guardian_name TEXT,
			date_of_birth TEXT,
			mobile_number TEXT NOT NULL,
			alternate_number TEXT,
			email TEXT,
			current_address TEXT,
			permanent_address TEXT,
			city_state_pin TEXT,
			city TEXT,
			state TEXT,
			pin TEXT,
			emergency_contact TEXT,
			partner_types TEXT[] NOT NULL DEFAULT '{}',
			service_categories TEXT[] NOT NULL DEFAULT '{}',
			primary_service TEXT,
			years_experience TEXT,
			preferred_locations TEXT,
			languages TEXT,
			skills TEXT,
			driving_licence TEXT,
			licence_expiry TEXT,
			vehicle_owner TEXT,
			vehicle_registration TEXT,
			vehicle_type TEXT,
			vehicle_make_model TEXT,
			rc_number TEXT,
			insurance_details TEXT,
			puc_valid_until TEXT,
			permit_details TEXT,
			business_name TEXT,
			business_type TEXT,
			authorized_person TEXT,
			business_address TEXT,
			gstin TEXT,
			pan TEXT,
			trade_registration TEXT,
			employees TEXT,
			team_size TEXT,
			years_in_business TEXT,
			previous_experience TEXT,
			government_id TEXT,
			pan_number TEXT,
			other_id_proof TEXT,
			bank_account_holder TEXT,
			bank_name TEXT,
			account_number TEXT,
			ifsc_code TEXT,
			upi_id TEXT,
			documents TEXT[] NOT NULL DEFAULT '{}',
			capabilities TEXT[] NOT NULL DEFAULT '{}',
			working_days TEXT,
			working_hours TEXT,
			service_radius TEXT,
			reference_one TEXT,
			reference_two TEXT,
			declaration_accepted BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			ip TEXT,
			status TEXT NOT NULL DEFAULT 'new'
		);
		CREATE INDEX IF NOT EXISTS idx_partner_registrations_created_at ON partner_registrations (created_at DESC);
		ALTER TABLE partner_registrations ADD COLUMN IF NOT EXISTS city TEXT;
		ALTER TABLE partner_registrations ADD COLUMN IF NOT EXISTS state TEXT;
		ALTER TABLE partner_registrations ADD COLUMN IF NOT EXISTS pin TEXT;
	`)
	return err
}

func handlePartnerRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var payload partnerRegistrationPayload
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256<<10)).Decode(&payload); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(payload.CompanyWebsite) != "" {
		respondJSON(w, http.StatusOK, map[string]bool{"success": true})
		return
	}

	payload.FullName = strings.TrimSpace(payload.FullName)
	payload.MobileNumber = strings.TrimSpace(payload.MobileNumber)
	payload.Email = strings.TrimSpace(payload.Email)
	payload.PrimaryService = strings.TrimSpace(payload.PrimaryService)
	if payload.PartnerTypes == nil {
		payload.PartnerTypes = []string{}
	}
	if payload.ServiceCategories == nil {
		payload.ServiceCategories = []string{}
	}
	if payload.Documents == nil {
		payload.Documents = []string{}
	}
	if payload.Capabilities == nil {
		payload.Capabilities = []string{}
	}
	if payload.FullName == "" || payload.MobileNumber == "" || payload.PrimaryService == "" || !payload.DeclarationAccepted {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "name, mobile number, primary service and declaration are required"})
		return
	}
	if len(payload.FullName) > 200 || len(payload.MobileNumber) > 40 || len(payload.Email) > 254 || len(payload.CurrentAddress) > 2000 || len(payload.BusinessAddress) > 2000 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "one or more fields are too long"})
		return
	}

	_, err := db.Exec(`
		INSERT INTO partner_registrations (
			full_name, guardian_name, date_of_birth, mobile_number, alternate_number, email,
			current_address, permanent_address, city, state, pin, emergency_contact, partner_types,
			service_categories, primary_service, years_experience, preferred_locations, languages,
			skills, driving_licence, licence_expiry, vehicle_owner, vehicle_registration, vehicle_type,
			vehicle_make_model, rc_number, insurance_details, puc_valid_until, permit_details,
			business_name, business_type, authorized_person, business_address, gstin, pan,
			trade_registration, employees, team_size, years_in_business, previous_experience,
			government_id, pan_number, other_id_proof, bank_account_holder, bank_name, account_number,
			ifsc_code, upi_id, documents, capabilities, working_days, working_hours, service_radius,
			reference_one, reference_two, declaration_accepted, ip
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,
			$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38,$39,$40,$41,$42,
			$43,$44,$45,$46,$47,$48,$49,$50,$51,$52,$53,$54,$55,$56,$57
		)`,
		payload.FullName, payload.GuardianName, payload.DateOfBirth, payload.MobileNumber, payload.AlternateNumber, payload.Email,
		payload.CurrentAddress, payload.PermanentAddress, payload.City, payload.State, payload.PIN, payload.EmergencyContact, pq.Array(payload.PartnerTypes),
		pq.Array(payload.ServiceCategories), payload.PrimaryService, payload.YearsExperience, payload.PreferredLocations, payload.Languages,
		payload.Skills, payload.DrivingLicence, payload.LicenceExpiry, payload.VehicleOwner, payload.VehicleRegistration, payload.VehicleType,
		payload.VehicleMakeModel, payload.RCNumber, payload.InsuranceDetails, payload.PUCValidUntil, payload.PermitDetails,
		payload.BusinessName, payload.BusinessType, payload.AuthorizedPerson, payload.BusinessAddress, payload.GSTIN, payload.PAN,
		payload.TradeRegistration, payload.Employees, payload.TeamSize, payload.YearsInBusiness, payload.PreviousExperience,
		payload.GovernmentID, payload.PANNumber, payload.OtherIDProof, payload.BankAccountHolder, payload.BankName, payload.AccountNumber,
		payload.IFSCCode, payload.UPIID, pq.Array(payload.Documents), pq.Array(payload.Capabilities), payload.WorkingDays, payload.WorkingHours,
		payload.ServiceRadius, payload.ReferenceOne, payload.ReferenceTwo, payload.DeclarationAccepted, clientIP(r),
	)
	if err != nil {
		log.Println("partner registration insert error:", err)
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not save registration, please try again"})
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func handleListPartnerRegistrations(adminKey string) http.HandlerFunc {
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
			SELECT id, COALESCE(full_name,''), COALESCE(guardian_name,''), COALESCE(date_of_birth,''),
			       COALESCE(mobile_number,''), COALESCE(alternate_number,''), COALESCE(email,''),
				   COALESCE(current_address,''), COALESCE(permanent_address,''), COALESCE(city,''), COALESCE(state,''),
				   COALESCE(pin,''), COALESCE(emergency_contact,''), partner_types, service_categories, COALESCE(primary_service,''),
			       COALESCE(years_experience,''), COALESCE(preferred_locations,''), COALESCE(languages,''),
			       COALESCE(skills,''), COALESCE(driving_licence,''), COALESCE(licence_expiry,''),
			       COALESCE(vehicle_owner,''), COALESCE(vehicle_registration,''), COALESCE(vehicle_type,''),
			       COALESCE(vehicle_make_model,''), COALESCE(rc_number,''), COALESCE(insurance_details,''),
			       COALESCE(puc_valid_until,''), COALESCE(permit_details,''), COALESCE(business_name,''),
			       COALESCE(business_type,''), COALESCE(authorized_person,''), COALESCE(business_address,''),
			       COALESCE(gstin,''), COALESCE(pan,''), COALESCE(trade_registration,''), COALESCE(employees,''),
			       COALESCE(team_size,''), COALESCE(years_in_business,''), COALESCE(previous_experience,''),
			       COALESCE(government_id,''), COALESCE(pan_number,''), COALESCE(other_id_proof,''),
			       COALESCE(bank_account_holder,''), COALESCE(bank_name,''), COALESCE(account_number,''),
			       COALESCE(ifsc_code,''), COALESCE(upi_id,''), documents, capabilities, COALESCE(working_days,''),
			       COALESCE(working_hours,''), COALESCE(service_radius,''), COALESCE(reference_one,''),
			       COALESCE(reference_two,''), declaration_accepted, created_at, COALESCE(ip,''), COALESCE(status,'new')
			FROM partner_registrations
			ORDER BY created_at DESC`)
		if err != nil {
			log.Println("partner registrations query error:", err)
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not read partner registrations"})
			return
		}
		defer rows.Close()

		registrations := []PartnerRegistration{}
		for rows.Next() {
			var p PartnerRegistration
			var partnerTypes, serviceCategories, documents, capabilities pq.StringArray
			var createdAt time.Time
			if err := rows.Scan(
				&p.ID, &p.FullName, &p.GuardianName, &p.DateOfBirth, &p.MobileNumber, &p.AlternateNumber, &p.Email,
				&p.CurrentAddress, &p.PermanentAddress, &p.City, &p.State, &p.PIN, &p.EmergencyContact, &partnerTypes, &serviceCategories,
				&p.PrimaryService, &p.YearsExperience, &p.PreferredLocations, &p.Languages, &p.Skills, &p.DrivingLicence,
				&p.LicenceExpiry, &p.VehicleOwner, &p.VehicleRegistration, &p.VehicleType, &p.VehicleMakeModel, &p.RCNumber,
				&p.InsuranceDetails, &p.PUCValidUntil, &p.PermitDetails, &p.BusinessName, &p.BusinessType, &p.AuthorizedPerson,
				&p.BusinessAddress, &p.GSTIN, &p.PAN, &p.TradeRegistration, &p.Employees, &p.TeamSize, &p.YearsInBusiness,
				&p.PreviousExperience, &p.GovernmentID, &p.PANNumber, &p.OtherIDProof, &p.BankAccountHolder, &p.BankName,
				&p.AccountNumber, &p.IFSCCode, &p.UPIID, &documents, &capabilities, &p.WorkingDays, &p.WorkingHours,
				&p.ServiceRadius, &p.ReferenceOne, &p.ReferenceTwo, &p.DeclarationAccepted, &createdAt, &p.IP, &p.Status,
			); err != nil {
				log.Println("partner registration scan error:", err)
				respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not read partner registrations"})
				return
			}
			p.PartnerTypes = []string(partnerTypes)
			p.ServiceCategories = []string(serviceCategories)
			p.Documents = []string(documents)
			p.Capabilities = []string(capabilities)
			p.CreatedAt = createdAt.UTC().Format(time.RFC3339)
			registrations = append(registrations, p)
		}
		if err := rows.Err(); err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not read partner registrations"})
			return
		}
		respondJSON(w, http.StatusOK, registrations)
	}
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
