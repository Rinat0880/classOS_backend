package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	classosbackend "github.com/rinat0880/classOS_backend"
	"github.com/rinat0880/classOS_backend/pkg/service"
)

type signInInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) signUp(c *gin.Context) {
	var input classosbackend.User

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.services.Authorization.CreateUser(input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})

}

func (h *Handler) signIn(c *gin.Context) {
	var input signInInput

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.services.Authorization.GenerateToken(input.Username, input.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			newErrorResponse(c, http.StatusUnauthorized, err.Error()) // 401
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, "something went wrong at server") // 500
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
	})
}
