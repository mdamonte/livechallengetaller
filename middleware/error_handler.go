package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandler is a middleware to handle errors across the application
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		errs := c.Errors
		if len(errors) == 0 {
			return
		}

		// Log the error
		for _, e := range errs {
			log.Printf("Error: %v", e.Err)
		}

		// Respond with the first error
		e := errs[0].Err
		switch e {
		case ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": e.Error()})
		case ErrInvalidInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
	}
}

// Custom error types
var (
	ErrNotFound     = NewError("resource not found")
	ErrInvalidInput = NewError("invalid input")
)

type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func NewError(message string) *Error {
	return &Error{Message: message}
}
