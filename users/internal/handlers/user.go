package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"users/internal/repository"
)

type UserHandler struct {
	repo userRepository
}

func NewUserHandler(repo userRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetString("userID")
	user, err := h.repo.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		log.Printf("[USERS] ERROR | get user: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID := c.GetString("userID")

	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"    binding:"omitempty,email"`
		Password string `json:"password" binding:"omitempty,min=8"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.Name == "" && body.Email == "" && body.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field is required"})
		return
	}

	// Fetch current user to fill missing fields for partial update.
	current, err := h.repo.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		log.Printf("[USERS] ERROR | get user: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if body.Name == "" {
		body.Name = current.Name
	}
	if body.Email == "" {
		body.Email = current.Email
	}

	if body.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("[USERS] ERROR | hash password: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if err := h.repo.UpdatePassword(c.Request.Context(), userID, string(hash)); err != nil {
			log.Printf("[USERS] ERROR | update password: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
	}

	user, err := h.repo.UpdateProfile(c.Request.Context(), userID, body.Name, body.Email)
	if err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already taken"})
			return
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		log.Printf("[USERS] ERROR | update user: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, user)
}
