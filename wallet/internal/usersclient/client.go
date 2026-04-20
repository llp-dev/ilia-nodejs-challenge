package usersclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrUserNotFound is returned when the users service responds with 404.
var ErrUserNotFound = errors.New("user not found")

type Client struct {
	baseURL    string
	httpClient *http.Client
	secret     []byte
}

func New(baseURL, secret string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		secret:     []byte(secret),
	}
}

func (c *Client) signedToken() (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "wallet-service",
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
	})
	return token.SignedString(c.secret)
}

// GetUser calls GET /users/:id on the users service and returns the user's email.
func (c *Client) GetUser(ctx context.Context, userID string) (string, error) {
	tok, err := c.signedToken()
	if err != nil {
		return "", fmt.Errorf("sign internal token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/users/"+userID, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tok)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", ErrUserNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d from users service", resp.StatusCode)
	}

	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return body.Email, nil
}
