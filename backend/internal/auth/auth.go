package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

const (
	SessionCookieName = "cerebray_session"
	SessionTTL        = 7 * 24 * time.Hour
	StateTTL          = 10 * time.Minute
)

// OIDCUserInfo holds standard OIDC claims from the userinfo endpoint.
type OIDCUserInfo struct {
	Subject string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// NewKeycloakOAuthConfig builds an OAuth2 config for Keycloak OIDC.
func NewKeycloakOAuthConfig(issuerURL, clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  issuerURL + "/protocol/openid-connect/auth",
			TokenURL: issuerURL + "/protocol/openid-connect/token",
		},
	}
}

// FetchOIDCUserInfo exchanges a token for user info from the Keycloak userinfo endpoint.
func FetchOIDCUserInfo(ctx context.Context, issuerURL string, token *oauth2.Token, oauthCfg *oauth2.Config) (*OIDCUserInfo, error) {
	client := oauthCfg.Client(ctx, token)
	resp, err := client.Get(issuerURL + "/protocol/openid-connect/userinfo")
	if err != nil {
		return nil, fmt.Errorf("fetching userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo returned status %d", resp.StatusCode)
	}

	var info OIDCUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decoding userinfo: %w", err)
	}
	return &info, nil
}

func GenerateRandomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func SetSessionCookie(w http.ResponseWriter, sessionID string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(SessionTTL.Seconds()),
	})
}

func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
