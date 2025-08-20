package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	classosbackend "github.com/rinat0880/classOS_backend"
)

func (h *Handler) createGroup(c *gin.Context) {
	userid, err := getUserId(c)
	if err != nil {
		return
	}

	var input classosbackend.Group
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.services.Group.Create(userid, input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})

}

type getAllGroupsResponse struct {
	Data []classosbackend.Group `json:"data"`
}

func (h *Handler) getAllGroups(c *gin.Context) {
	userid, err := getUserId(c)
	if err != nil {
		return
	}

	groups, err := h.services.Group.GetAll(userid)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, getAllGroupsResponse{
		Data: groups,
	})
}

func (h *Handler) getGroupById(c *gin.Context) {
	userid, err := getUserId(c)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id in params")
		return
	}

	group, err := h.services.Group.GetById(userid, id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, group)

}

func (h *Handler) updateGroup(c *gin.Context) {

}

func (h *Handler) deleteGroup(c *gin.Context) {

}