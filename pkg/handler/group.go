package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	classosbackend "github.com/rinat0880/classOS_backend"
)

func (h *Handler) createGroup(c *gin.Context) {
	checkerId, err := getUserId(c)
	if err != nil {
		return
	}

	var input classosbackend.Group
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.services.Group.Create(checkerId, input)
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
	checkerId, err := getUserId(c)
	if err != nil {
		return
	}

	groups, err := h.services.Group.GetAll(checkerId)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, getAllGroupsResponse{
		Data: groups,
	})
}

func (h *Handler) getGroupById(c *gin.Context) {
	checkerId, err := getUserId(c)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id in params")
		return
	}

	group, err := h.services.Group.GetById(checkerId, id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, group)

}

func (h *Handler) updateGroup(c *gin.Context) {
	checkerId, err := getUserId(c)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id in params")
		return
	}

	var input classosbackend.UpdateGroupInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}


	if err := h.services.Group.Update(checkerId, id, input); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{
		Status: "Ok",
	})
}

func (h *Handler) deleteGroup(c *gin.Context) {
	checkerId, err := getUserId(c)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id in params")
		return
	}

	err = h.services.Group.Delete(checkerId, id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{
		Status: "ok",
	})
}