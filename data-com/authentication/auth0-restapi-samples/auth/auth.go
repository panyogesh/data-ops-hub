package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/auth0-restapi-samples/config"
)

type AuthConfig struct {
	Domain string
	API    struct {
		Audience string
	}
	M2M struct {
		ClientID     string
		ClientSecret string
	}
	Web struct {
		ClientID     string
		ClientSecret string
		RedirectURI  string
	}
}

func randState(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func NewAuth(config config.Config) *AuthConfig {
	return &AuthConfig{
		Domain: config.Auth.Domain,
		API: struct {
			Audience string
		}{
			Audience: config.Auth.API.Audience,
		},
		M2M: struct {
			ClientID     string
			ClientSecret string
		}{
			ClientID:     config.Auth.M2M.ClientID,
			ClientSecret: config.Auth.M2M.ClientSecret,
		},
		Web: struct {
			ClientID     string
			ClientSecret string
			RedirectURI  string
		}{
			ClientID:     config.Auth.Web.ClientID,
			RedirectURI:  config.Auth.Web.RedirectURI,
			ClientSecret: config.Auth.Web.ClientSecret,
		},
	}
}

func (a *AuthConfig) LoginServices() string {
	// Define your parameters
	clientID := a.Web.ClientID
	redirectURI := a.Web.RedirectURI
	baseURL := "https://" + a.Domain + "/authorize"

	// Using url.Values to safely encode the query string
	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", clientID)
	params.Add("redirect_uri", redirectURI)
	params.Add("audience", a.API.Audience)

	fmt.Println("redirectURI:", redirectURI)

	// IMPORTANT for Universal Login / OIDC
	params.Set("scope", "openid profile email")
	params.Set("state", randState(16))
	params.Set("response_mode", "query") // optional but harmless

	// Construct final URL: https://<URL>/authorize?client_id=...
	finalURL := baseURL + "?" + params.Encode()

	fmt.Println("Final URL:", finalURL)

	return finalURL
}

func (a *AuthConfig) TokenValidation(code string) (int, map[string]interface{}) {
	// Define your parameters
	clientID := a.Web.ClientID
	redirectURI := a.Web.RedirectURI

	from := url.Values{}
	from.Set("grant_type", "authorization_code")
	from.Set("client_id", clientID)
	from.Set("client_secret", a.Web.ClientSecret)
	from.Set("code", code)
	from.Set("redirect_uri", redirectURI)

	tokenURL := "https://" + a.Domain + "/oauth/token"

	req, err := http.NewRequest("POST", tokenURL, io.NopCloser(strings.NewReader(from.Encode())))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return http.StatusInternalServerError, map[string]interface{}{"error": err.Error()}
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return http.StatusInternalServerError, map[string]interface{}{"error": err.Error()}
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, map[string]interface{}{"error": string(raw)}
	}

	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		// If body isn't JSON for some reason, return raw
		return resp.StatusCode, map[string]interface{}{"data": string(raw)}
	}

	return resp.StatusCode, map[string]interface{}{"data": out}
}
