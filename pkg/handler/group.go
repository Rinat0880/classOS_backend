package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) createGroup(c *gin.Context) {
	id, _ := c.Get(userCtx)
	c.JSON(http.StatusOK, map[string]interface{} {
		"id": id,
	}) 
}

func (h *Handler) getAllGroups(c *gin.Context) {

}

func (h *Handler) getGroupById(c *gin.Context) {

}

func (h *Handler) updateGroup(c *gin.Context) {

}

func (h *Handler) deleteGroup(c *gin.Context) {

}