package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/karan-singh-17/Quick-Mail/database"
	"github.com/karan-singh-17/Quick-Mail/models"
)

// CurrentUser handles requests to retrieve the current user's information.
// It checks for a valid JWT token in the cookies and extracts the user's details.
// This function requires authentication and returns the current user's information if the token is valid.
// @Summary Retrieve the current user's information
// @Description Checks the provided JWT token in the cookies for validity, and retrieves the user's details from the database.
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} string "User information retrieved successfully"
// @Failure 401 {string} string "Unauthorized: Invalid or missing JWT token"
// @Failure 404 {string} string "User not found"
// @Router /api/user/current [get]
func CurrentUser(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tokenString := cookie.Value
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-256-bit-secret"), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", claims.Issuer).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"status": http.StatusOK,
		"user":   user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
