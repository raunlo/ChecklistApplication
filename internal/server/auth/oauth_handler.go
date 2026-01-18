package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxResponseBodySize = 1 << 20 // 1MB limit for HTTP responses

// googleHTTPClient is a shared HTTP client with proper timeouts for Google API calls
var googleHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	},
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// NewGoogleOAuthConfig creates and validates a new GoogleOAuthConfig
func NewGoogleOAuthConfig(clientID, clientSecret, redirectURL string) (*GoogleOAuthConfig, error) {
	if clientID == "" {
		return nil, errors.New("client ID is required")
	}
	if clientSecret == "" {
		return nil, errors.New("client secret is required")
	}
	if redirectURL == "" {
		return nil, errors.New("redirect URL is required")
	}

	return &GoogleOAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
	}, nil
}

type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	IdToken      string `json:"id_token"`
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// GetAuthURL generates the Google OAuth authorization URL
func (c *GoogleOAuthConfig) GetAuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", c.ClientID)
	params.Set("redirect_uri", c.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", "openid email profile")
	params.Set("access_type", "offline")
	params.Set("prompt", "consent")
	params.Set("state", state)

	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

// ExchangeCode exchanges authorization code for tokens
func (c *GoogleOAuthConfig) ExchangeCode(ctx context.Context, code string) (*GoogleTokenResponse, error) {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	data.Set("redirect_uri", c.RedirectURL)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token",
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := googleHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
		if err != nil {
			return nil, fmt.Errorf("token exchange failed with status %d: failed to read response body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, body)
	}

	var tokens GoogleTokenResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBodySize)).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	// Validate required fields
	if tokens.AccessToken == "" {
		return nil, errors.New("access token missing in response")
	}

	return &tokens, nil
}

// GetUserInfo fetches user information from Google
func (c *GoogleOAuthConfig) GetUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := googleHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
		if err != nil {
			return nil, fmt.Errorf("user info request failed with status %d: failed to read response body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("user info request failed with status %d: %s", resp.StatusCode, body)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBodySize)).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info response: %w", err)
	}

	// Validate required fields
	if userInfo.ID == "" {
		return nil, errors.New("user ID missing in response")
	}
	if userInfo.Email == "" {
		return nil, errors.New("user email missing in response")
	}

	return &userInfo, nil
}

// RefreshAccessToken uses refresh token to get a new access token
func (c *GoogleOAuthConfig) RefreshAccessToken(ctx context.Context, refreshToken string) (*GoogleTokenResponse, error) {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	data := url.Values{}
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token",
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := googleHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
		if err != nil {
			return nil, fmt.Errorf("token refresh failed with status %d: failed to read response body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, body)
	}

	var tokens GoogleTokenResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBodySize)).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	// Validate required fields
	if tokens.AccessToken == "" {
		return nil, errors.New("access token missing in response")
	}

	return &tokens, nil
}
