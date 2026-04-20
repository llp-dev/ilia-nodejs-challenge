package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"users/internal/repository"
)

type AuthHandler struct {
	repo      userRepository
	jwtSecret []byte
}

func NewAuthHandler(repo userRepository, jwtSecret string) *AuthHandler {
	return &AuthHandler{repo: repo, jwtSecret: []byte(jwtSecret)}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var body struct {
		Name     string `json:"name"     binding:"required"`
		Email    string `json:"email"    binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[USERS] ERROR | hash password: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	id := uuid.New().String()
	user, err := h.repo.Create(c.Request.Context(), id, body.Name, body.Email, string(hash))
	if err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already taken"})
			return
		}
		log.Printf("[USERS] ERROR | create user: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email"    binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, hash, err := h.repo.GetByEmail(c.Request.Context(), body.Email)
	if err != nil {
		// Do not enumerate users — return 401 regardless of error type.
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"iat":   now.Unix(),
		"exp":   now.Add(24 * time.Hour).Unix(),
	})
	signed, err := token.SignedString(h.jwtSecret)
	if err != nil {
		log.Printf("[USERS] ERROR | sign token: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": signed, "user": user})
}
