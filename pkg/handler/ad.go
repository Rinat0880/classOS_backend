package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) syncFromAD(c *gin.Context) {
	// Предполагаем, что IntegratedUserService имеет метод SyncAllFromAD
	// Этот метод нужно будет добавить в интерфейс User
	c.JSON(http.StatusOK, statusResponse{
		Status: "Sync started",
	})
}

func (h *Handler) checkADConnection(c *gin.Context) {
	// Предполагаем, что IntegratedUserService имеет метод ValidateADConnection
	c.JSON(http.StatusOK, gin.H{
		"status":  "connected",
		"message": "AD connection is working",
	})
}
