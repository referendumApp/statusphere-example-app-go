package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/bluesky-social/statusphere-go/internal/config"
	"github.com/bluesky-social/statusphere-go/internal/db"
	"github.com/bluesky-social/statusphere-go/internal/view"
	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"
)

// Handlers holds all HTTP handlers
type Handlers struct {
	cfg      *config.Config
	db       *db.DB
	store    *sessions.CookieStore
	templates *template.Template
}

// New creates a new Handlers instance
func New(cfg *config.Config, database *db.DB) *Handlers {
	// Create cookie store for sessions
	store := sessions.NewCookieStore([]byte(cfg.CookieSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30, // 30 days
		HttpOnly: true,
	}

	// Load templates
	tmpl := template.Must(template.ParseGlob(filepath.Join("templates", "*.html")))

	return &Handlers{
		cfg:       cfg,
		db:        database,
		store:     store,
		templates: tmpl,
	}
}

// ClientMetadata serves OAuth client metadata
func (h *Handlers) ClientMetadata(w http.ResponseWriter, r *http.Request) {
	publicURL := h.cfg.PublicURL
	if publicURL == "" {
		publicURL = "http://" + r.Host
	}

	// Create client metadata similar to the original NodeOAuthClient
	metadata := map[string]interface{}{
		"client_name": "AT Protocol Express App (Go)",
		"client_id":   publicURL + "/client-metadata.json",
		"client_uri":  publicURL,
		"redirect_uris": []string{
			publicURL + "/oauth/callback",
		},
		"scope":                     "atproto transition:generic",
		"grant_types":               []string{"authorization_code", "refresh_token"},
		"response_types":            []string{"code"},
		"application_type":          "web",
		"token_endpoint_auth_method": "none",
		"dpop_bound_access_tokens":  true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}

// OAuthCallback handles the OAuth callback
func (h *Handlers) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	// To be implemented when we integrate AT Protocol client
	http.Redirect(w, r, "/?error=not_implemented", http.StatusFound)
}

// ShowLogin displays the login page
func (h *Handlers) ShowLogin(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Error": "",
	}

	view.RenderTemplate(w, "login", data)
}

// HandleLogin processes login form submission
func (h *Handlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// To be implemented when we integrate AT Protocol client
	http.Redirect(w, r, "/?error=not_implemented", http.StatusFound)
}

// HandleLogout processes logout
func (h *Handlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Get session
	session, _ := h.store.Get(r, "sid")

	// Clear session values
	session.Values = map[interface{}]interface{}{}

	// Save session
	if err := session.Save(r, w); err != nil {
		log.Error().Err(err).Msg("Failed to save session")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// Home displays the homepage
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	// Get statuses from database
	statuses, err := h.db.GetRecentStatuses(10)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get statuses")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get session
	session, _ := h.store.Get(r, "sid")

	// Get user DID from session if logged in
	userDID, ok := session.Values["did"].(string)

	var myStatus *db.Status
	var profile map[string]interface{}

	// If user is logged in, get their status
	if ok && userDID != "" {
		var err error
		myStatus, err = h.db.GetUserStatus(userDID)
		if err != nil {
			log.Debug().Err(err).Msg("User has no status")
			// This is not a critical error, user might not have a status yet
		}

		// For now, just create a minimal profile
		// This will be replaced with actual profile data in the AT Protocol integration
		profile = map[string]interface{}{
			"displayName": session.Values["displayName"],
		}
	}

	// For now, just use a simple map for the DID to handle mapping
	// This will be replaced with actual handle resolution in the AT Protocol integration
	didHandleMap := make(map[string]string)
	for _, status := range statuses {
		didHandleMap[status.AuthorDID] = status.AuthorDID
	}

	data := map[string]interface{}{
		"Statuses":     statuses,
		"DidHandleMap": didHandleMap,
		"Profile":      profile,
		"MyStatus":     myStatus,
	}

	view.RenderTemplate(w, "home", data)
}

// UpdateStatus updates the user's status
func (h *Handlers) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	// Get session
	session, _ := h.store.Get(r, "sid")

	// Get user DID from session
	userDID, ok := session.Values["did"].(string)
	if !ok || userDID == "" {
		http.Error(w, "Error: Session required", http.StatusUnauthorized)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error: Invalid form data", http.StatusBadRequest)
		return
	}

	// Get status from form
	statusText := r.FormValue("status")
	if statusText == "" {
		http.Error(w, "Error: Invalid status", http.StatusBadRequest)
		return
	}

	// Create status
	// This is a simplified implementation, will be replaced with actual AT Protocol integration
	now := time.Now().UTC().Format(time.RFC3339)
	status := &db.Status{
		URI:       "at://" + userDID + "/xyz.statusphere.status/temporary",
		AuthorDID: userDID,
		Status:    statusText,
		CreatedAt: now,
		IndexedAt: now,
	}

	// Save status to database
	if err := h.db.SaveStatus(status); err != nil {
		log.Error().Err(err).Msg("Failed to save status")
		http.Error(w, "Error: Failed to save status", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}