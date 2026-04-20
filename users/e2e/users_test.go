package e2e_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"users/internal/models"
	"users/internal/testhelper"
)

func TestUsersLifecycle(t *testing.T) {
	testhelper.Truncate(t, testPool)

	// 1. Register
	regBody, _ := json.Marshal(map[string]string{
		"name":     "Alice",
		"email":    "alice@example.com",
		"password": "password123",
	})
	resp, err := http.DefaultClient.Do(req(http.MethodPost, testServer.URL+"/users", bytes.NewReader(regBody), ""))
	if err != nil {
		t.Fatalf("POST /users: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /users status = %d, want 201", resp.StatusCode)
	}

	var user models.User
	json.NewDecoder(resp.Body).Decode(&user)
	if user.ID == "" {
		t.Fatal("expected non-empty user ID")
	}

	// 2. Duplicate registration returns 409
	resp2, _ := http.DefaultClient.Do(req(http.MethodPost, testServer.URL+"/users", bytes.NewReader(regBody), ""))
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusConflict {
		t.Errorf("duplicate register status = %d, want 409", resp2.StatusCode)
	}

	// 3. Login
	loginBody, _ := json.Marshal(map[string]string{
		"email":    "alice@example.com",
		"password": "password123",
	})
	resp3, err := http.DefaultClient.Do(req(http.MethodPost, testServer.URL+"/sessions", bytes.NewReader(loginBody), ""))
	if err != nil {
		t.Fatalf("POST /sessions: %v", err)
	}
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("POST /sessions status = %d, want 200", resp3.StatusCode)
	}

	var loginResp map[string]interface{}
	json.NewDecoder(resp3.Body).Decode(&loginResp)
	token, _ := loginResp["token"].(string)
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// 4. Wrong password returns 401
	badBody, _ := json.Marshal(map[string]string{
		"email":    "alice@example.com",
		"password": "wrongpass",
	})
	resp4, _ := http.DefaultClient.Do(req(http.MethodPost, testServer.URL+"/sessions", bytes.NewReader(badBody), ""))
	resp4.Body.Close()
	if resp4.StatusCode != http.StatusUnauthorized {
		t.Errorf("bad password status = %d, want 401", resp4.StatusCode)
	}

	// 5. GET /users/me
	resp5, err := http.DefaultClient.Do(req(http.MethodGet, testServer.URL+"/users/me", nil, token))
	if err != nil {
		t.Fatalf("GET /users/me: %v", err)
	}
	defer resp5.Body.Close()

	if resp5.StatusCode != http.StatusOK {
		t.Fatalf("GET /users/me status = %d, want 200", resp5.StatusCode)
	}

	var meUser models.User
	json.NewDecoder(resp5.Body).Decode(&meUser)
	if meUser.ID != user.ID {
		t.Errorf("me.ID = %q, want %q", meUser.ID, user.ID)
	}

	// 6. PUT /users/me — update name
	updateBody, _ := json.Marshal(map[string]string{"name": "Alicia"})
	resp6, err := http.DefaultClient.Do(req(http.MethodPut, testServer.URL+"/users/me", bytes.NewReader(updateBody), token))
	if err != nil {
		t.Fatalf("PUT /users/me: %v", err)
	}
	defer resp6.Body.Close()

	if resp6.StatusCode != http.StatusOK {
		t.Fatalf("PUT /users/me status = %d, want 200", resp6.StatusCode)
	}

	var updated models.User
	json.NewDecoder(resp6.Body).Decode(&updated)
	if updated.Name != "Alicia" {
		t.Errorf("updated name = %q, want Alicia", updated.Name)
	}

	// 7. GET /users/me without token returns 401
	resp7, _ := http.DefaultClient.Do(req(http.MethodGet, testServer.URL+"/users/me", nil, ""))
	resp7.Body.Close()
	if resp7.StatusCode != http.StatusUnauthorized {
		t.Errorf("no-token status = %d, want 401", resp7.StatusCode)
	}
}
