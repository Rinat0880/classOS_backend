package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	classosbackend "github.com/rinat0880/classOS_backend"
)

func (h *Handler) createUser(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		return
	}

	groupId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id in params")
		return
	}

	var input classosbackend.User
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.services.User.Create(userId, groupId, input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})
}

func (h *Handler) getAllUsers(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		return
	}

	groupId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id in params")
		return
	}

	users, err := h.services.User.GetAll(userId, groupId)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *Handler) getUserById(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		return
	}

	user_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id in params")
		return
	}

	users, err := h.services.User.GetAll(userId, user_id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *Handler) updateUser(c *gin.Context) {

}

func (h *Handler) deleteUser(c *gin.Context) {

}
