package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	classosbackend "github.com/rinat0880/classOS_backend"
)

func (h *Handler) createUser(c *gin.Context) {
	checkerId, err := getUserId(c)
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

	id, err := h.services.User.Create(checkerId, groupId, input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})
}

func (h *Handler) getAllUsers(c *gin.Context) {
	checkerId, err := getUserId(c)
	if err != nil {
		return
	}

	users, err := h.services.User.GetAll(checkerId)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *Handler) getUserById(c *gin.Context) {
	checkerId, err := getUserId(c)
	if err != nil {
		return
	}

	user_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id in params")
		return
	}

	user, err := h.services.User.GetById(checkerId, user_id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) updateUser(c *gin.Context) {
	checkerId, err := getUserId(c)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id in params")
		return
	}

	var input classosbackend.UpdateUserInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}


	if err := h.services.User.Update(checkerId, id, input); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{
		Status: "Ok",
	})
}

func (h *Handler) deleteUser(c *gin.Context) {
	checkerId, err := getUserId(c)
	if err != nil {
		return
	}

	user_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id in params")
		return
	}

	err = h.services.User.Delete(checkerId, user_id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{"ok"})
}

func (h *Handler) changePassword(c *gin.Context) {
	checkerId, err := getUserId(c)
	if err != nil {
		return
	}

	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id in params")
		return
	}

	var input struct {
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	updateInput := classosbackend.UpdateUserInput{
		Password: &input.NewPassword,
	}

	if err := h.services.User.Update(checkerId, userId, updateInput); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{
		Status: "Password changed successfully",
	})
}
