package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/karan-singh-17/Quick-Mail/database"
)

var templates = template.Must(template.ParseFiles("templates/verify_success.html", "templates/verify_fail.html"))

// VerifyUser handles the verification of a user based on a token provided in the URL path.
// It checks if the token exists in a temporary store, creates the user in the database if the token is valid,
// and then responds based on the 'Accept' header in the request.
// - If the 'Accept' header specifies 'application/json', it returns a JSON response.
// - If the 'Accept' header specifies 'text/html', it renders an HTML template.
// @Summary Verify user based on token
// @Description This endpoint verifies a user by checking a token from the URL path. If the token is valid and exists in the temporary store, the user is created in the database. The response format depends on the 'Accept' header in the request. JSON is returned if the header contains 'application/json'; otherwise, an HTML template is rendered.
// @Tags User Verification
// @Accept json
// @Produce json
// @Param token path string true "Verification Token"
// @Success 200 {object} string "User verified and created successfully"
// @Failure 400 {string} string "Invalid or expired token"
// @Failure 500 {string} string "Internal Server Error: Failed to create user"
// @Router /api/user/verify/{token} [get]
func VerifyUser(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Path[len("/api/user/verify/"):]

	tempStore.Lock()
	user, exists := tempStore.data[token]
	if exists {
		delete(tempStore.data, token)
	}
	tempStore.Unlock()

	if !exists {
		log.Println("Invalid or expired token")
		renderError(w, http.StatusBadRequest, "Invalid or expired token")
		return
	}

	if err := database.DB.Create(&user).Error; err != nil {
		log.Println("Error creating user:", err)
		renderError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	response := map[string]interface{}{
		"status":  http.StatusOK,
		"message": "User verified and created successfully",
	}
	/*
		//w.Header().Set("Content-Type", "application/json")
		//json.NewEncoder(w).Encode(response)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := templates.ExecuteTemplate(w, "verify_success.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}*/

	acceptHeader := r.Header.Get("Accept")

	if strings.Contains(acceptHeader, "application/json") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	} else {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := templates.ExecuteTemplate(w, "verify_success.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
func renderError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := templates.ExecuteTemplate(w, "verify_fail.html", map[string]string{"Message": message}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
