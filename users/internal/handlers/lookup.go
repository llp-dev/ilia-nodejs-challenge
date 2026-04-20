package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"users/internal/repository"
)

// LookupHandler serves the internal GET /users/:id endpoint used by the wallet service.
type LookupHandler struct {
	repo userRepository
}

func NewLookupHandler(repo userRepository) *LookupHandler {
	return &LookupHandler{repo: repo}
}

func (h *LookupHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	user, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		log.Printf("[USERS] ERROR | lookup user: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	// Return only id and email as agreed with the wallet service.
	c.JSON(http.StatusOK, gin.H{"id": user.ID, "email": user.Email})
}
