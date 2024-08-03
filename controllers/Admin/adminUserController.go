package controllers

import (
	"mountgear/helpers"
	"mountgear/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ........................list all users with details.........................................
func ListUsers(ctx *gin.Context) {
	var users []models.User

	if err := models.FetchData(models.DB, &users); err != nil {
		helpers.SendResponse(ctx, http.StatusInternalServerError, "Could not fetch users", nil)
	}
	helpers.SendResponse(ctx, http.StatusOK, "Success", nil, gin.H{"users": users})
}

//.........................................Block user.............................................

func BlockUser(c *gin.Context) {
	var user models.User

	if err := models.FindUserByID(models.DB, c.Param("id"), &user); err != nil {
		helpers.SendResponse(c, http.StatusNotFound, "Could not find user", nil)
		return
	}

	user.IsActive = false

	if err := models.UpdateRecord(models.DB, &user); err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Could not update user", nil)
	}

	helpers.SendResponse(c, http.StatusOK, "User Block Successfully", nil)

}

// ......................................Unblock user...............................................
func UnBlockUser(c *gin.Context) {
	var user models.User

	if err := models.FindUserByID(models.DB, c.Param("id"), &user); err != nil {
		helpers.SendResponse(c, http.StatusNotFound, "Could not find user", nil)
		return
	}

	user.IsActive = true // Unblock the user

	if err := models.UpdateRecord(models.DB, &user); err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Faild to update user", nil)
		return
	}

	helpers.SendResponse(c, http.StatusOK, "User unlock successfully", nil)

}
