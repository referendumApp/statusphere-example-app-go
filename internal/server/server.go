package server

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/referendumApp/statusphere-example-app-go/internal/config"
	"github.com/referendumApp/statusphere-example-app-go/internal/db"
	"github.com/referendumApp/statusphere-example-app-go/internal/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// Server represents the HTTP server
type Server struct {
	cfg        *config.Config
	db         *db.DB
	router     *mux.Router
	httpServer *http.Server
}

// New creates a new server instance
func New(cfg *config.Config, database *db.DB) (*Server, error) {
	s := &Server{
		cfg:    cfg,
		db:     database,
		router: mux.NewRouter(),
	}

	// Initialize the server
	if err := s.initialize(); err != nil {
		return nil, err
	}

	return s, nil
}

// initialize sets up the HTTP routes and middleware
func (s *Server) initialize() error {
	// Create the handlers with dependencies
	h := handlers.New(s.cfg, s.db)

	// Set up middleware
	s.router.Use(loggingMiddleware)
	s.router.Use(recoveryMiddleware)

	// Static assets (using the existing front-end assets)
	fs := http.FileServer(http.Dir(filepath.Join(".", "static")))
	s.router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", fs))

	// OAuth routes
	s.router.HandleFunc("/client-metadata.json", h.ClientMetadata).Methods("GET")
	s.router.HandleFunc("/oauth/callback", h.OAuthCallback).Methods("GET")

	// Authentication routes
	s.router.HandleFunc("/login", h.ShowLogin).Methods("GET")
	s.router.HandleFunc("/login", h.HandleLogin).Methods("POST")
	s.router.HandleFunc("/logout", h.HandleLogout).Methods("POST")

	// Main routes
	s.router.HandleFunc("/", h.Home).Methods("GET")
	s.router.HandleFunc("/status", h.UpdateStatus).Methods("POST")

	// 404 handler
	s.router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "404 Not Found")
	})

	// Create the HTTP server
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Middleware functions

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log the request
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Dur("duration", time.Since(start)).
			Msg("Request processed")
	})
}

// recoveryMiddleware recovers from panics
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Interface("error", err).
					Msg("Panic recovered")

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}