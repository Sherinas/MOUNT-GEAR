// package main

// import (
// 	"log"
// 	"mountgear/middlewares"
// 	"mountgear/models"
// 	"mountgear/routes"
// 	"mountgear/utils"
// 	"net/http"
// 	"os"

// 	"github.com/gin-gonic/gin"
// 	"github.com/joho/godotenv"
// 	"golang.org/x/oauth2"
// 	"golang.org/x/oauth2/google"
// )

// func init() {
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal("Error loading .env file")
// 	}

// 	utils.GoogleOauthConfig = &oauth2.Config{
// 		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
// 		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
// 		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
// 		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
// 		Endpoint:     google.Endpoint,
// 	}
// }

// func main() {

// 	// ensureUploadsDir()
// 	db := models.DatabaseSetup()
// 	if db == nil {
// 		log.Fatalf("Failed to connect to database")
// 	}
// 	models.SetDatabase(db)
// 	models.AutoMigrate(db)
// 	// RunMigrations()

// 	r := gin.Default()
// 	r.Use(gin.Logger())
// 	r.Use(middlewares.NoCacheMiddleware())

// 	//r.Static("/assets", "./assets")

// 	r.LoadHTMLGlob("templates/*")

// 	routes.AuthRoutes(r)
// 	routes.HomeRoutes(r)
// 	routes.AdminRoutes(r)
// 	routes.OAuthRoutes(r)
// 	auth := r.Group("/")
// 	auth.Use(middlewares.AuthMiddleware())
// 	{
// 		auth.GET("/protected", func(c *gin.Context) {
// 			userID, _ := c.Get("userID")
// 			c.JSON(http.StatusOK, gin.H{"user_id": userID})
// 		})
// 	}

// 	r.Run(":3030")

//		//.................................................
//	}
package main

import (
	"log"
	"mountgear/middlewares"
	"mountgear/models"
	"mountgear/routes"
	"mountgear/utils"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	utils.GoogleOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

func main() {
	// Database setup

	db := models.DatabaseSetup()
	if db == nil {
		log.Fatalf("Failed to connect to database")
	}
	models.SetDatabase(db)
	models.AutoMigrate(db)
	Migrate(db)

	// Initialize Gin router
	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(middlewares.NoCacheMiddleware())

	// Static files and templates
	r.LoadHTMLGlob("templates/*")

	// Define routes
	routes.AuthRoutes(r)
	routes.HomeRoutes(r)
	routes.AdminRoutes(r)
	routes.OAuthRoutes(r)

	// Run the server
	if err := r.Run(":3030"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
func Migrate(db *gorm.DB) {
	db.AutoMigrate(&models.Order{})
}
