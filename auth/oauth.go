package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/digiveljet/netatmo-cli/config"
)

const (
	authURL  = "https://api.netatmo.com/oauth2/authorize"
	tokenURL = "https://api.netatmo.com/oauth2/token"
	scope    = "read_station"
)

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
}

// Login starts OAuth2 authorization code flow with a local callback server.
func Login(clientID, clientSecret string) (*config.Config, error) {
	// Find available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errMsg := r.URL.Query().Get("error")
			fmt.Fprintf(w, "<html><body><h2>Authorization failed: %s</h2></body></html>", errMsg)
			errCh <- fmt.Errorf("authorization failed: %s", errMsg)
			return
		}
		fmt.Fprintf(w, "<html><body><h2>✅ Authorized! You can close this tab.</h2></body></html>")
		codeCh <- code
	})

	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(listener) }()

	// Build authorization URL
	params := url.Values{
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"scope":         {scope},
		"state":         {"netatmo-cli"},
		"response_type": {"code"},
	}
	authLink := authURL + "?" + params.Encode()

	fmt.Println("Open this URL in your browser to authorize:")
	fmt.Println()
	fmt.Println(authLink)
	fmt.Println()
	fmt.Println("Waiting for authorization...")

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		return nil, err
	case <-time.After(2 * time.Minute):
		return nil, fmt.Errorf("authorization timed out")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)

	// Exchange code for tokens
	cfg, err := exchangeCode(clientID, clientSecret, code, redirectURI)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func exchangeCode(clientID, clientSecret, code, redirectURI string) (*config.Config, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"scope":         {scope},
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tok tokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}
	if tok.Error != "" {
		return nil, fmt.Errorf("token error: %s", tok.Error)
	}

	return &config.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
	}, nil
}

// Refresh gets a new access token using the refresh token.
func Refresh(cfg *config.Config) error {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"refresh_token": {cfg.RefreshToken},
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return fmt.Errorf("refresh failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tok tokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return fmt.Errorf("failed to parse refresh response: %w", err)
	}
	if tok.Error != "" {
		return fmt.Errorf("refresh error: %s — run 'netatmo auth' again", tok.Error)
	}

	cfg.AccessToken = tok.AccessToken
	if tok.RefreshToken != "" {
		cfg.RefreshToken = tok.RefreshToken
	}
	return cfg.Save()
}

// EnsureToken loads config and refreshes token if needed.
func EnsureToken() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	// Try to use current token first; refresh on 401/403
	resp, err := http.Get("https://api.netatmo.com/api/getstationsdata?access_token=" + cfg.AccessToken)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		if err := Refresh(cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

// AuthenticatedGet makes a GET request with Bearer token, auto-refreshing if needed.
func AuthenticatedGet(cfg *config.Config, endpoint string) ([]byte, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		if err := Refresh(cfg); err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return body, nil
}
