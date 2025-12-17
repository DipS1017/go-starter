package apple

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AppleConfig represents configuration for Sign-in with Apple
type AppleConfig struct {
	TeamID      string
	ClientID    string
	RedirectURI string
	KeyID       string
	PrivateKey  any
}

// AppleAuthToken represents Apple's token response
type AppleAuthToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

// Constants for Apple's OAuth endpoints
const (
	AppleAuthURL  = "https://appleid.apple.com/auth/token"
	AppleAudience = "https://appleid.apple.com"
	AppleGrant    = "authorization_code"
)

// LoadPrivateKey loads an ES256 private key from a byte array (PEM format)
func (a *AppleConfig) LoadPrivateKey(key []byte) error {
	block, _ := pem.Decode(key)
	if block == nil {
		return errors.New("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse PKCS8 private key: %w", err)
	}

	a.PrivateKey = privateKey
	return nil
}

// LoadPrivateKeyFromFile loads an ES256 private key from a file
func (a *AppleConfig) LoadPrivateKeyFromFile(path string) error {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}
	return a.LoadPrivateKey(keyData)
}

// NewAppleConfig creates a new AppleConfig instance
func NewAppleConfig(teamID, clientID, redirectURI, keyID string) *AppleConfig {
	return &AppleConfig{
		TeamID:      teamID,
		ClientID:    clientID,
		RedirectURI: redirectURI,
		KeyID:       keyID,
	}
}

// CreateCallbackURL generates Apple's authorization URL
func (a *AppleConfig) CreateCallbackURL(state string) string {
	params := url.Values{
		"response_type": {"code"},
		"redirect_uri":  {a.RedirectURI},
		"client_id":     {a.ClientID},
		"state":         {state},
		"scope":         {"openid"},
	}
	return "https://appleid.apple.com/auth/authorize?" + params.Encode()
}

// generateClientSecret creates a JWT client secret for Apple OAuth
func (a *AppleConfig) generateClientSecret(expiration time.Duration) (string, error) {
	if a.PrivateKey == nil {
		return "", errors.New("private key not loaded")
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": a.TeamID,
		"iat": now.Unix(),
		"exp": now.Add(expiration).Unix(),
		"aud": AppleAudience,
		"sub": a.ClientID,
	})

	token.Header["kid"] = a.KeyID

	clientSecret, err := token.SignedString(a.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return clientSecret, nil
}

// GetAppleToken exchanges an authorization code for an AppleAuthToken
func (a *AppleConfig) GetAppleToken(code string, clientSecretExpiry time.Duration) (string, error) {
	clientSecret, err := a.generateClientSecret(clientSecretExpiry)
	if err != nil {
		return "", err
	}

	formData := url.Values{
		"client_id":     {a.ClientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"grant_type":    {AppleGrant},
		"redirect_uri":  {a.RedirectURI},
	}

	resp, err := http.PostForm(AppleAuthURL, formData)
	if err != nil {
		return "", fmt.Errorf("request to Apple failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("apple returned non-200 status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var authToken AppleAuthToken
	if err := json.NewDecoder(resp.Body).Decode(&authToken); err != nil {
		return "", fmt.Errorf("failed to parse Apple token response: %w", err)
	}

	return authToken.IDToken, nil
}
