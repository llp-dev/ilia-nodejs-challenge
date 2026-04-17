package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"wallet/internal/handlers"
	"wallet/internal/models"
)

func TestWalletHandler_List_DBError(t *testing.T) {
	r := gin.New()
	repo := &mockWalletRepo{
		listFn: func(ctx context.Context) ([]models.Wallet, error) {
			return nil, errDB
		},
	}
	h := handlers.NewWalletHandler(repo)
	r.GET("/wallets", h.List)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/wallets", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestWalletHandler_Create_DBError(t *testing.T) {
	r := gin.New()
	repo := &mockWalletRepo{
		createFn: func(ctx context.Context, userID, description string) (*models.Wallet, error) {
			return nil, errDB
		},
	}
	h := handlers.NewWalletHandler(repo)
	r.POST("/wallets", h.Create)

	body := []byte(`{"user_id":"550e8400-e29b-41d4-a716-446655440000"}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestTransactionHandler_Create_DBError(t *testing.T) {
	r := gin.New()
	repo := &mockTransactionRepo{
		createFn: func(ctx context.Context, walletID string, value decimal.Decimal, description, operationID string) (*models.Transaction, error) {
			return nil, errDB
		},
	}
	h := handlers.NewTransactionHandler(repo)
	r.POST("/wallets/:id/transactions", h.Create)

	body := []byte(`{"value":"10.00","description":"test","operation_id":"550e8400-e29b-41d4-a716-446655440000"}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/wallets/some-id/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
