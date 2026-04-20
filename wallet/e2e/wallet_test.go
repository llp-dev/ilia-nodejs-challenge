package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"wallet/internal/models"
	"wallet/internal/testhelper"
)

func TestWalletLifecycle(t *testing.T) {
	testhelper.Truncate(t, testPool)

	// 1. Create a wallet
	wallet := createWalletE2E(t, "550e8400-e29b-41d4-a716-446655440000", "my wallet")
	if wallet.ID == "" {
		t.Fatal("expected wallet ID")
	}
	if wallet.Description != "my wallet" {
		t.Errorf("description = %q, want %q", wallet.Description, "my wallet")
	}

	// 2. List wallets — should contain the created one
	wallets := listWalletsE2E(t)
	if len(wallets) != 1 {
		t.Fatalf("expected 1 wallet, got %d", len(wallets))
	}

	// 3. Get wallet by ID
	found := getWalletE2E(t, wallet.ID)
	if found.ID != wallet.ID {
		t.Errorf("ID = %q, want %q", found.ID, wallet.ID)
	}

	// 4. Update description
	updated := updateDescriptionE2E(t, wallet.ID, "updated description")
	if updated.Description != "updated description" {
		t.Errorf("description = %q, want %q", updated.Description, "updated description")
	}

	// 5. Credit the wallet
	postTransactionE2E(t, wallet.ID, "100.00", "deposit", "550e8400-e29b-41d4-a716-000000000001", http.StatusCreated)

	// 6. Verify balance
	after := getWalletE2E(t, wallet.ID)
	if after.Balance.String() != "100" {
		t.Errorf("balance = %s, want 100", after.Balance.String())
	}

	// 7. Debit the wallet
	postTransactionE2E(t, wallet.ID, "-40.00", "withdrawal", "550e8400-e29b-41d4-a716-000000000002", http.StatusCreated)

	after = getWalletE2E(t, wallet.ID)
	if after.Balance.String() != "60" {
		t.Errorf("balance = %s, want 60", after.Balance.String())
	}

	// 8. Overdraft should be rejected
	postTransactionE2E(t, wallet.ID, "-999.00", "overdraft", "550e8400-e29b-41d4-a716-000000000003", http.StatusUnprocessableEntity)

	// Balance must remain unchanged
	after = getWalletE2E(t, wallet.ID)
	if after.Balance.String() != "60" {
		t.Errorf("balance = %s after rejected overdraft, want 60", after.Balance.String())
	}
}

func TestWallet_Ownership(t *testing.T) {
	testhelper.Truncate(t, testPool)

	// Create a wallet belonging to testUserID via the normal flow.
	wallet := createWalletE2E(t, testUserID, "my wallet")

	// Build a second JWT for a different user.
	otherID := "00000000-0000-0000-0000-aaaaaaaaaaaa"
	otherToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   otherID,
		"email": "other@example.com",
	}).SignedString([]byte(testSecret))
	otherAuth := func(method, url string, body interface{ Read([]byte) (int, error) }) *http.Request {
		var req *http.Request
		if body != nil {
			req, _ = http.NewRequest(method, url, body)
		} else {
			req, _ = http.NewRequest(method, url, nil)
		}
		req.Header.Set("Authorization", "Bearer "+otherToken)
		req.Header.Set("Content-Type", "application/json")
		return req
	}

	t.Run("get wallet owned by another user returns 403", func(t *testing.T) {
		resp, err := http.DefaultClient.Do(otherAuth(http.MethodGet, testServer.URL+"/wallets/"+wallet.ID, nil))
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("status = %d, want 403", resp.StatusCode)
		}
	})

	t.Run("update wallet owned by another user returns 403", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"description": "hijack"})
		resp, err := http.DefaultClient.Do(otherAuth(http.MethodPut, testServer.URL+"/wallets/"+wallet.ID, bytes.NewReader(body)))
		if err != nil {
			t.Fatalf("PUT: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("status = %d, want 403", resp.StatusCode)
		}
	})

	t.Run("transaction on wallet owned by another user returns 403", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"value": "10.00", "description": "hijack", "operation_id": "550e8400-e29b-41d4-a716-000000000099",
		})
		resp, err := http.DefaultClient.Do(otherAuth(http.MethodPost, testServer.URL+"/wallets/"+wallet.ID+"/transactions", bytes.NewReader(body)))
		if err != nil {
			t.Fatalf("POST: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("status = %d, want 403", resp.StatusCode)
		}
	})

	t.Run("list returns only own wallets", func(t *testing.T) {
		resp, err := http.DefaultClient.Do(otherAuth(http.MethodGet, testServer.URL+"/wallets", nil))
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		var wallets []models.Wallet
		json.NewDecoder(resp.Body).Decode(&wallets)
		if len(wallets) != 0 {
			t.Errorf("expected 0 wallets for other user, got %d", len(wallets))
		}
	})
}

func TestWallet_NotFound(t *testing.T) {
	testhelper.Truncate(t, testPool)

	t.Run("get non-existent wallet returns 404", func(t *testing.T) {
		resp, err := http.DefaultClient.Do(authReq(http.MethodGet, testServer.URL+"/wallets/00000000-0000-0000-0000-000000000000", nil))
		if err != nil {
			t.Fatalf("GET wallet: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("status = %d, want 404", resp.StatusCode)
		}
	})

	t.Run("transaction on non-existent wallet returns 404", func(t *testing.T) {
		postTransactionE2E(t, "00000000-0000-0000-0000-000000000000", "10.00", "ghost", "550e8400-e29b-41d4-a716-000000000099", http.StatusNotFound)
	})
}

// --- helpers ---

func createWalletE2E(t *testing.T, userID, description string) models.Wallet {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"user_id": userID, "description": description})
	resp, err := http.DefaultClient.Do(authReq(http.MethodPost, testServer.URL+"/wallets", bytes.NewReader(body)))
	if err != nil {
		t.Fatalf("POST /wallets: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /wallets status = %d, want 201", resp.StatusCode)
	}
	var w models.Wallet
	json.NewDecoder(resp.Body).Decode(&w)
	return w
}

func listWalletsE2E(t *testing.T) []models.Wallet {
	t.Helper()
	resp, err := http.DefaultClient.Do(authReq(http.MethodGet, testServer.URL+"/wallets", nil))
	if err != nil {
		t.Fatalf("GET /wallets: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /wallets status = %d, want 200", resp.StatusCode)
	}
	var wallets []models.Wallet
	json.NewDecoder(resp.Body).Decode(&wallets)
	return wallets
}

func getWalletE2E(t *testing.T, id string) models.Wallet {
	t.Helper()
	resp, err := http.DefaultClient.Do(authReq(http.MethodGet, testServer.URL+"/wallets/"+id, nil))
	if err != nil {
		t.Fatalf("GET /wallets/%s: %v", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /wallets/%s status = %d, want 200", id, resp.StatusCode)
	}
	var w models.Wallet
	json.NewDecoder(resp.Body).Decode(&w)
	return w
}

func updateDescriptionE2E(t *testing.T, id, description string) models.Wallet {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"description": description})
	resp, err := http.DefaultClient.Do(authReq(http.MethodPut, testServer.URL+"/wallets/"+id, bytes.NewReader(body)))
	if err != nil {
		t.Fatalf("PUT /wallets/%s: %v", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("PUT /wallets/%s status = %d, want 200", id, resp.StatusCode)
	}
	var w models.Wallet
	json.NewDecoder(resp.Body).Decode(&w)
	return w
}

func postTransactionE2E(t *testing.T, walletID, value, description, operationID string, wantStatus int) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"value":        value,
		"description":  description,
		"operation_id": operationID,
	})
	resp, err := http.DefaultClient.Do(authReq(
		http.MethodPost,
		fmt.Sprintf("%s/wallets/%s/transactions", testServer.URL, walletID),
		bytes.NewReader(body),
	))
	if err != nil {
		t.Fatalf("POST /wallets/%s/transactions: %v", walletID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != wantStatus {
		t.Errorf("POST /wallets/%s/transactions status = %d, want %d", walletID, resp.StatusCode, wantStatus)
	}
}
