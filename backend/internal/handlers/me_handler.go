package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"userId": c.GetString("userId"),
		"role":   c.GetString("role"),
	})
}
