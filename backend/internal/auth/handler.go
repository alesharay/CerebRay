package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

// UserUpsertFunc creates or updates a user from OIDC claims and returns the user ID.
type UserUpsertFunc func(ctx context.Context, oidcSubject, email, name, avatar string) (int64, error)

// UserGetFunc fetches a user by ID. Returns the user as an interface to avoid circular deps with sqlc.
type UserGetFunc func(ctx context.Context, id int64) (interface{}, error)

// Handler serves auth endpoints.
type Handler struct {
	oauthCfg     *oauth2.Config
	issuerURL    string
	sessions     *SessionStore
	upsertUser   UserUpsertFunc
	getUser      UserGetFunc
	baseURL      string
	secureCookie bool
	localMode    bool
	localEmail   string
	localName    string
}

type HandlerConfig struct {
	OAuthConfig  *oauth2.Config
	IssuerURL    string
	Sessions     *SessionStore
	UpsertUser   UserUpsertFunc
	GetUser      UserGetFunc
	BaseURL      string
	SecureCookie bool
	LocalMode    bool
	LocalEmail   string
	LocalName    string
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		oauthCfg:     cfg.OAuthConfig,
		issuerURL:    cfg.IssuerURL,
		sessions:     cfg.Sessions,
		upsertUser:   cfg.UpsertUser,
		getUser:      cfg.GetUser,
		baseURL:      cfg.BaseURL,
		secureCookie: cfg.SecureCookie,
		localMode:    cfg.LocalMode,
		localEmail:   cfg.LocalEmail,
		localName:    cfg.LocalName,
	}
}

// HandleLogin redirects to Keycloak or auto-logs in for local mode.
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if h.localMode {
		h.handleLocalLogin(w, r)
		return
	}

	state, err := GenerateRandomToken(16)
	if err != nil {
		log.Error().Err(err).Msg("generating state token")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	verifier := oauth2.GenerateVerifier()

	if err := h.sessions.StoreStateToken(r.Context(), state, verifier); err != nil {
		log.Error().Err(err).Msg("storing state token")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	url := h.oauthCfg.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleCallback processes the OIDC callback.
func (h *Handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	if h.localMode {
		http.Redirect(w, r, h.baseURL+"/dashboard", http.StatusTemporaryRedirect)
		return
	}

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if state == "" || code == "" {
		http.Error(w, "missing state or code", http.StatusBadRequest)
		return
	}

	verifier, err := h.sessions.ValidateAndConsumeStateToken(r.Context(), state)
	if err != nil {
		log.Error().Err(err).Msg("invalid state token")
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	token, err := h.oauthCfg.Exchange(r.Context(), code, oauth2.VerifierOption(verifier))
	if err != nil {
		log.Error().Err(err).Msg("exchanging code for token")
		http.Error(w, "auth exchange failed", http.StatusInternalServerError)
		return
	}

	userInfo, err := FetchOIDCUserInfo(r.Context(), h.issuerURL, token, h.oauthCfg)
	if err != nil {
		log.Error().Err(err).Msg("fetching user info")
		http.Error(w, "failed to get user info", http.StatusInternalServerError)
		return
	}

	userID, err := h.upsertUser(r.Context(), userInfo.Subject, userInfo.Email, userInfo.Name, userInfo.Picture)
	if err != nil {
		log.Error().Err(err).Msg("upserting user")
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	sessionID, err := h.sessions.Create(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("creating session")
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	SetSessionCookie(w, sessionID, h.secureCookie)
	http.Redirect(w, r, h.baseURL+"/dashboard", http.StatusTemporaryRedirect)
}

// HandleLogout destroys the session and clears the cookie.
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err == nil {
		_ = h.sessions.Delete(r.Context(), cookie.Value)
	}
	ClearSessionCookie(w, h.secureCookie)
	w.WriteHeader(http.StatusNoContent)
}

// HandleMe returns the current user info.
// Expects GetUserIDFromContext to be set before use.
var GetUserIDFromContext func(ctx context.Context) int64

func (h *Handler) HandleMe(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.getUser(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("fetching user for /me")
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) handleLocalLogin(w http.ResponseWriter, r *http.Request) {
	userID, err := h.upsertUser(r.Context(), "local", h.localEmail, h.localName, "")
	if err != nil {
		log.Error().Err(err).Msg("creating local user")
		http.Error(w, "failed to create local user", http.StatusInternalServerError)
		return
	}

	sessionID, err := h.sessions.Create(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("creating session")
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	SetSessionCookie(w, sessionID, h.secureCookie)
	http.Redirect(w, r, h.baseURL+"/dashboard", http.StatusTemporaryRedirect)
}
