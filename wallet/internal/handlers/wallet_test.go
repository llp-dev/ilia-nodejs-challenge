package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"wallet/internal/handlers"
	"wallet/internal/models"
	"wallet/internal/repository"
	"wallet/internal/testhelper"
)

const (
	testUserID    = "550e8400-e29b-41d4-a716-446655440000"
	testUserEmail = "user@example.com"
)

func newWalletRouter() *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
		c.Set("userEmail", testUserEmail)
		c.Next()
	})
	walletRepo := repository.NewWalletRepository(testPool)
	usersClient := &mockUsersClient{
		getUserFn: func(ctx context.Context, userID string) (string, error) {
			return testUserEmail, nil
		},
	}
	h := handlers.NewWalletHandler(walletRepo, usersClient)
	r.GET("/wallets", h.List)
	r.GET("/wallets/:id", h.GetByID)
	r.POST("/wallets", h.Create)
	r.PUT("/wallets/:id", h.UpdateDescription)
	return r
}

func createWallet(t *testing.T, userID, description string) models.Wallet {
	t.Helper()
	repo := repository.NewWalletRepository(testPool)
	w, err := repo.Create(context.Background(), userID, description)
	if err != nil {
		t.Fatalf("seed wallet: %v", err)
	}
	return *w
}

func TestWalletHandler_List(t *testing.T) {
	testhelper.Truncate(t, testPool)
	r := newWalletRouter()

	t.Run("returns empty list when no wallets", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/wallets", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
		var result []models.Wallet
		json.NewDecoder(w.Body).Decode(&result)
		if len(result) != 0 {
			t.Errorf("got %d wallets, want 0", len(result))
		}
	})

	t.Run("returns only wallets owned by authenticated user", func(t *testing.T) {
		createWallet(t, testUserID, "wallet one")
		createWallet(t, testUserID, "wallet two")
		createWallet(t, "550e8400-e29b-41d4-a716-446655440001", "other user wallet")

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/wallets", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
		var result []models.Wallet
		json.NewDecoder(w.Body).Decode(&result)
		if len(result) != 2 {
			t.Errorf("got %d wallets, want 2", len(result))
		}
	})
}

func TestWalletHandler_GetByID(t *testing.T) {
	testhelper.Truncate(t, testPool)
	r := newWalletRouter()
	wallet := createWallet(t, testUserID, "my wallet")

	t.Run("returns wallet by id", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/wallets/"+wallet.ID, nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
		var result models.Wallet
		json.NewDecoder(w.Body).Decode(&result)
		if result.ID != wallet.ID {
			t.Errorf("ID = %q, want %q", result.ID, wallet.ID)
		}
	})

	t.Run("returns 403 for wallet owned by another user", func(t *testing.T) {
		other := createWallet(t, "550e8400-e29b-41d4-a716-446655440001", "other user wallet")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/wallets/"+other.ID, nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("returns 404 for non-existent wallet", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/wallets/00000000-0000-0000-0000-000000000000", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestWalletHandler_Create(t *testing.T) {
	testhelper.Truncate(t, testPool)
	r := newWalletRouter()

	t.Run("creates wallet", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"user_id":     testUserID,
			"description": "new wallet",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
		}
		var result models.Wallet
		json.NewDecoder(w.Body).Decode(&result)
		if result.ID == "" {
			t.Error("expected non-empty ID")
		}
		if result.UserID != testUserID {
			t.Errorf("UserID = %q, want %q", result.UserID, testUserID)
		}
	})

	t.Run("returns 400 when user_id is missing", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"description": "no user",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestWalletHandler_UpdateDescription(t *testing.T) {
	testhelper.Truncate(t, testPool)
	r := newWalletRouter()
	wallet := createWallet(t, testUserID, "original")

	t.Run("updates description", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"description": "updated"})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/wallets/"+wallet.ID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
		var result models.Wallet
		json.NewDecoder(w.Body).Decode(&result)
		if result.Description != "updated" {
			t.Errorf("Description = %q, want %q", result.Description, "updated")
		}
	})

	t.Run("returns 400 when description is missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/wallets/"+wallet.ID, bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("returns 400 for unknown fields", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"description": "x", "unknown": "field"})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/wallets/"+wallet.ID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/wallets/"+wallet.ID, bytes.NewReader([]byte(`not-json`)))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("returns 403 for wallet owned by another user", func(t *testing.T) {
		other := createWallet(t, "550e8400-e29b-41d4-a716-446655440001", "other user")
		body, _ := json.Marshal(map[string]string{"description": "hijack"})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/wallets/"+other.ID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("returns 404 for non-existent wallet", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"description": "x"})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/wallets/00000000-0000-0000-0000-000000000000", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
