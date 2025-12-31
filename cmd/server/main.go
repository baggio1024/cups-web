package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/gorilla/mux"
    "cups-web/frontend"
    "cups-web/internal/ipp"
    "cups-web/internal/auth"
    "cups-web/internal/middleware"
    "cups-web/internal/server"
)

func main() {
    addr := os.Getenv("LISTEN_ADDR")
    if addr == "" {
        addr = ":8080"
    }

    hashKey := os.Getenv("SESSION_HASH_KEY")
    blockKey := os.Getenv("SESSION_BLOCK_KEY")
    if hashKey == "" || blockKey == "" {
        log.Println("Warning: SESSION_HASH_KEY or SESSION_BLOCK_KEY not set; using generated keys (not recommended for production)")
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

    // Static files (embedded) - register after API routes so /api/* is matched first
    serverFS := server.NewEmbeddedServer(frontend.FS)
    r.PathPrefix("/").Handler(serverFS)

    srv := &http.Server{
        Addr:         addr,
        Handler:      r,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
    }

    fmt.Println("listening on", addr)
    log.Fatal(srv.ListenAndServe())
}
