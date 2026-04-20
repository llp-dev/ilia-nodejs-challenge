package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"wallet/internal/repository"
)

type WalletHandler struct {
	repo walletRepository
}

func NewWalletHandler(repo walletRepository) *WalletHandler {
	return &WalletHandler{repo: repo}
}

func (h *WalletHandler) List(c *gin.Context) {
	wallets, err := h.repo.List(c.Request.Context())
	if err != nil {
		log.Printf("[WALLET] ERROR | List wallets: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, wallets)
}

func (h *WalletHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	wallet, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}
	c.JSON(http.StatusOK, wallet)
}

func (h *WalletHandler) Create(c *gin.Context) {
	var body struct {
		UserID      string `json:"user_id" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wallet, err := h.repo.Create(c.Request.Context(), body.UserID, body.Description)
	if err != nil {
		log.Printf("[WALLET] ERROR | Create wallet: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, wallet)
}

func (h *WalletHandler) UpdateDescription(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Description string `json:"description"`
	}
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		msg := "invalid request body"
		if strings.Contains(err.Error(), "unknown field") {
			msg = "unknown fields are not allowed"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	if body.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description is required"})
		return
	}

	wallet, err := h.repo.UpdateDescription(c.Request.Context(), id, body.Description)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}
	c.JSON(http.StatusOK, wallet)
}

type TransactionHandler struct {
	repo transactionRepository
}

func NewTransactionHandler(repo transactionRepository) *TransactionHandler {
	return &TransactionHandler{repo: repo}
}

func (h *TransactionHandler) Create(c *gin.Context) {
	walletID := c.Param("id")
	var body struct {
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
		OperationID string `json:"operation_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	value, err := decimal.NewFromString(body.Value)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value"})
		return
	}

	t, err := h.repo.Create(c.Request.Context(), walletID, value, body.Description, body.OperationID)
	if err != nil {
		if errors.Is(err, repository.ErrInsufficientBalance) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, repository.ErrWalletNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		log.Printf("[WALLET] ERROR | Create transaction: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, t)
}
