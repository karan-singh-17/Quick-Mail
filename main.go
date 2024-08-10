package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/karan-singh-17/Quick-Mail/database"
	"github.com/karan-singh-17/Quick-Mail/routes"
	"github.com/rs/cors"
)

// @title Quick Mail API
// @version 1.0
// @Description The Quick Mail API allows users to create groups, add recipients manually or from a CSV file (either from a device or a link), and send messages (plain text, HTML from a link, or an HTML file) to multiple recipients in a single API call. This API is designed to simplify mass email communications by providing flexible options for managing recipient lists and message content. \n ### Use Cases:\n1. **Group Creation and Email Campaigns**: A user can create a group and manually add recipients or import them from a CSV file. The user can then send a message to all members of the group. \n2. **Dynamic Email Campaigns**: If a user has a dynamic list of recipients stored online in a CSV file, they can use the file's URL to add recipients to the group without manually updating the list. \n3. **HTML Email Campaigns**: Users can send HTML emails either by uploading an HTML file or by providing a link to an HTML template hosted online. \n4. **Simple Text Campaigns**: Users can quickly send plain text messages to a list of recipients without the need for HTML formatting. \n ### Security: \nThe Quick Mail API is secure and incorporates a custom-built two-factor authentication (2FA) system. Upon user registration, a verification email is sent to confirm the user's email address. When logging in, a 6-digit code is sent to the user's email, which must be entered to gain access. This ensures that only authorized users can create groups and send messages. Also API's are incorporated with a jwt_token authentication which allows only active users to manage their groups and hence is secured.
// @contact.name Karan Singh
// @contact.email karansingh122134@gmail.com
// @contact.url https://github.com/karan-singh-17/Quick_Mail_server
// @host https://quickmailserver-production.up.railway.app
// @BasePath /
// @securityDefinitions.jwt_token
// @securityDefinitions.jwt_token.type apiKey
// @securityDefinitions.jwt_token.name Cookie
// @securityDefinitions.jwt_token.in header
// @securityDefinitions.jwt_token.description "JWT token for authentication"
func main() {

	//err := godotenv.Load()
	//if err != nil {
	//	log.Fatalf("Error loading .env file")
	//}

	database.Connect()

	mux := http.NewServeMux()
	routes.SetupRoutes(mux)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		Debug:            false,
	})

	handler := c.Handler(mux)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
