package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"cups-web/frontend"
	"cups-web/internal/auth"
	"cups-web/internal/ipp"
	"cups-web/internal/middleware"
	"cups-web/internal/server"
	"cups-web/internal/store"
	"github.com/gorilla/mux"
)

func main() {
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = filepath.Join("data", "cups-web.db")
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Fatal("failed to create data dir: ", err)
	}
	var err error
	appStore, err = store.Open(context.Background(), dbPath)
	if err != nil {
		log.Fatal("failed to open database: ", err)
	}
	if err := ensureDefaultAdmin(context.Background()); err != nil {
		log.Fatal("failed to ensure default admin: ", err)
	}

	uploadDir = os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal("failed to create uploads dir: ", err)
	}

	hashKey := os.Getenv("SESSION_HASH_KEY")
	blockKey := os.Getenv("SESSION_BLOCK_KEY")
	if hashKey == "" || blockKey == "" {
		log.Println("Warning: SESSION_HASH_KEY or SESSION_BLOCK_KEY not set; using generated keys (not recommended for production). Set them via environment variables to keep sessions stable.")
	}
	auth.SetupSecureCookie(hashKey, blockKey)

	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/login", LoginHandler).Methods("POST")
	api.HandleFunc("/logout", LogoutHandler).Methods("POST")
	api.HandleFunc("/csrf", CSRFHandler).Methods("GET")
	// session endpoint used by frontend to detect existing session on page load
	api.HandleFunc("/session", SessionHandler).Methods("GET")

	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.RequireSession)
	protected.Use(middleware.ValidateCSRF)
	protected.HandleFunc("/me", MeHandler).Methods("GET")
	protected.HandleFunc("/printers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		cupsHost := os.Getenv("CUPS_HOST")
		if cupsHost == "" {
			cupsHost = "localhost"
		}

		printers, err := ipp.ListPrinters(cupsHost)
		if err != nil {
			http.Error(w, "failed to list printers: "+err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(printers)
	}).Methods("GET")
	protected.HandleFunc("/print", printHandler).Methods("POST")
	protected.HandleFunc("/convert", convertHandler).Methods("POST")
	protected.HandleFunc("/estimate", estimateHandler).Methods("POST")
	protected.HandleFunc("/print-records", printRecordsHandler).Methods("GET")
	protected.HandleFunc("/print-records/{id:[0-9]+}/file", printRecordFileHandler).Methods("GET")

	admin := api.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.RequireSession)
	admin.Use(middleware.RequireAdmin)
	admin.Use(middleware.ValidateCSRF)
	admin.HandleFunc("/users", adminListUsersHandler).Methods("GET")
	admin.HandleFunc("/users", adminCreateUserHandler).Methods("POST")
	admin.HandleFunc("/users/{id:[0-9]+}", adminUpdateUserHandler).Methods("PUT")
	admin.HandleFunc("/users/{id:[0-9]+}", adminDeleteUserHandler).Methods("DELETE")
	admin.HandleFunc("/users/{id:[0-9]+}/topup", adminTopupHandler).Methods("POST")
	admin.HandleFunc("/print-records", adminPrintRecordsHandler).Methods("GET")
	admin.HandleFunc("/topups", adminTopupsHandler).Methods("GET")
	admin.HandleFunc("/settings", adminGetSettingsHandler).Methods("GET")
	admin.HandleFunc("/settings", adminUpdateSettingsHandler).Methods("PUT")

	// Static files (embedded) - register after API routes so /api/* is matched first
	serverFS := server.NewEmbeddedServer(frontend.FS)
	r.PathPrefix("/").Handler(serverFS)

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	startMaintenance(appStore, uploadDir)

	fmt.Println("listening on", addr)
	log.Fatal(srv.ListenAndServe())
}
