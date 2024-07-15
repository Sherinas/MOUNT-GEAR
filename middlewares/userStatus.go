package middlewares

// func UserStatus() gin.HandlerFunc {
// 	return func(c *gin.Context) {

// 		token, err := c.Cookie("token")
// 		if err != nil || token == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized",
// 				"message": "please login first"})

// 			c.Abort()
// 			return
// 		}
// 		claims, err := utils.ValidateToken(token)
// 		if err != nil {
// 			// If the token is invalid, clear the cookie and redirect to login
// 			c.SetCookie("token", "", -1, "/", "localhost", false, true)
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error":   "unauthorized",
// 				"message": "please login first"})

// 			c.Abort()
// 			return
// 		}
// 		var user models.User
// 		if err := models.DB.First(&user, claims.UserID).Error; err != nil {
// 			// If user not found, clear the cookie
// 			c.SetCookie("token", "", -1, "/", "localhost", false, true)
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error":   "unauthorized",
// 				"message": "please login first"})
// 			c.Abort()
// 			return
// 		}

// 		// Check if the user is active
// 		if !user.IsActive {

// 			c.SetCookie("token", "", -1, "/", "localhost", false, true)
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized",
// 				"message": "your access is temporarily blocked "})

// 			c.Abort()
// 			return
// 		}

// 		c.Next()
// 	}
// }
