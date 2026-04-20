package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"wallet/internal/handlers"
	"wallet/internal/models"
	"wallet/internal/repository"
	"wallet/internal/testhelper"
)

func newTransactionRouter() *gin.Engine {
	r := gin.New()
	txRepo := repository.NewTransactionRepository(testPool)
	h := handlers.NewTransactionHandler(txRepo)
	r.POST("/wallets/:id/transactions", h.Create)
	return r
}

func seedBalance(t *testing.T, walletID string, amount string) {
	t.Helper()
	txRepo := repository.NewTransactionRepository(testPool)
	_, err := txRepo.Create(context.Background(), walletID, decimal.RequireFromString(amount), "seed", "00000000-0000-0000-0000-000000000099")
	if err != nil {
		t.Fatalf("seed balance: %v", err)
	}
}

func TestTransactionHandler_Create(t *testing.T) {
	testhelper.Truncate(t, testPool)
	r := newTransactionRouter()
	wallet := createWallet(t, "550e8400-e29b-41d4-a716-446655440000", "test wallet")

	t.Run("credits wallet", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"value":        "50.00",
			"description":  "deposit",
			"operation_id": "550e8400-e29b-41d4-a716-446655440010",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+wallet.ID+"/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
		}
		var result models.Transaction
		json.NewDecoder(w.Body).Decode(&result)
		if result.ID == "" {
			t.Error("expected non-empty transaction ID")
		}
	})

	t.Run("debits wallet", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"value":        "-20.00",
			"description":  "withdrawal",
			"operation_id": "550e8400-e29b-41d4-a716-446655440011",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+wallet.ID+"/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
		}
	})

	t.Run("returns 422 on insufficient balance", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"value":        "-9999.00",
			"description":  "overdraft",
			"operation_id": "550e8400-e29b-41d4-a716-446655440012",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+wallet.ID+"/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
		}
	})

	t.Run("returns 404 for non-existent wallet", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"value":        "10.00",
			"description":  "ghost",
			"operation_id": "550e8400-e29b-41d4-a716-446655440013",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/wallets/00000000-0000-0000-0000-000000000000/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("returns 400 when value is missing", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"description":  "no value",
			"operation_id": "550e8400-e29b-41d4-a716-446655440014",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+wallet.ID+"/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("returns 400 when operation_id is missing", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"value":       "10.00",
			"description": "no op id",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+wallet.ID+"/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("returns 409 for duplicate operation_id", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"value":        "10.00",
			"description":  "duplicate",
			"operation_id": "550e8400-e29b-41d4-a716-446655440010",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+wallet.ID+"/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
		}
	})

	t.Run("returns 400 for invalid value format", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"value":        "not-a-number",
			"description":  "bad value",
			"operation_id": "550e8400-e29b-41d4-a716-446655440015",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+wallet.ID+"/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}
