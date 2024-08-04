package controllers

import (
	"mountgear/helpers"
	"mountgear/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ........................list all users with details.........................................
// func ListUsers(ctx *gin.Context) {
// 	var users []models.User

// 	type userdata struct {
// 		ID       int    `json:"id"`
// 		Name     string `json:"name"`
// 		Email    string `json:"email"`
// 		Phone    string `json:"phone"`
// 		IsActive bool   `json:"is_active"`
// 	}
// 	if err := models.FetchData(models.DB, &users); err != nil {
// 		helpers.SendResponse(ctx, http.StatusInternalServerError, "Could not fetch users", nil)

// 	}

//		helpers.SendResponse(ctx, http.StatusOK, "Success", nil, gin.H{"users": users})
//	}
func ListUsers(ctx *gin.Context) {
	var users []models.User

	type userdata struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		IsActive bool   `json:"is_active"`
	}

	if err := models.FetchData(models.DB, &users); err != nil {
		helpers.SendResponse(ctx, http.StatusInternalServerError, "Could not fetch users", nil)
		return
	}

	var userResponses []userdata
	for _, user := range users {
		userResponses = append(userResponses, userdata{
			ID:       int(user.ID),
			Name:     user.Name,
			Email:    user.Email,
			Phone:    user.Phone,
			IsActive: user.IsActive,
		})
	}

	helpers.SendResponse(ctx, http.StatusOK, "", nil, gin.H{"respose": userResponses})
}

//.........................................Block user.............................................

func BlockUser(c *gin.Context) {
	var user models.User

	if err := models.FindUserByID(models.DB, c.Param("id"), &user); err != nil {
		helpers.SendResponse(c, http.StatusNotFound, "Could not find user", nil)
		return
	}

	// if err := models.UpdateRecord(models.DB, &user); err != nil {
	// 	helpers.SendResponse(c, http.StatusInternalServerError, "Could not update user", nil)
	// }

	if err := models.DB.Model(&user).Where("id = ?", user.ID).Update("is_active", false).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Could not update user", nil)
		return
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

	if err := models.DB.Model(&user).Where("id = ?", user.ID).Update("is_active", true).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Could not update user", nil)
		return
	}

	helpers.SendResponse(c, http.StatusOK, "User unlock successfully", nil)

}
