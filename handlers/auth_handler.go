package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/karan-singh-17/Quick-Mail/database"
	"github.com/karan-singh-17/Quick-Mail/models"
	"golang.org/x/crypto/bcrypt"
)

// RegisterUser handles the registration of a new user. It processes the request to register a user by:
// - Checking if the request method is POST.
// - Parsing the request body to extract user data (email and password).
// - Verifying if the email is already registered.
// - Generating a unique verification token and hashing the password.
// - Storing the user details in a temporary store for verification.
// - Sending a verification email to the user with the token.
// - Returning a response indicating the registration status.
// @Summary Register a new user
// @Description This endpoint allows a new user to register by providing an email and password. The email is checked for existing registration, and a verification token is generated and sent via email. The user must verify their email to complete the registration process.
// @Tags User
// @Accept json
// @Produce json
// @Param user body map[string]string true "User Registration Data"
// @Success 201 {object} string "User registered. Please check your email to verify your account." "User registration successful"
// @Failure 400 {string} string "Invalid request body"
// @Failure 409 {string} string "Email is already registered"
// @Failure 500 {string} string "Internal server error"
// @Router /api/user/register [post]
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data map[string]string
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := data["email"]

	var existingUser models.User

	if err := database.DB.Where("email = ?", email).First(&existingUser).Error; err == nil {
		http.Error(w, "Email is already registered", http.StatusConflict)
		return
	}

	token, err := generateToken()
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data["password"]), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user := models.User{
		Id:       GenerateID(data["email"]),
		Email:    data["email"],
		Password: string(hashedPassword),
	}

	tempStore.Lock()
	tempStore.data[token] = user
	tempStore.Unlock()

	if err := sendVerificationEmail(user.Email, token); err != nil {
		log.Println("Error sending verification email:", err)
		http.Error(w, "Failed to send verification email", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":  http.StatusCreated,
		"message": "User registered. Please check your email to verify your account.",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
	log.Println("User registration initiated, verification link generated.")
}

// Login handles user authentication by validating email and password,
// and sends a login code to the user's email for further verification.
// It expects a POST request with JSON body containing email and password.
// On successful authentication, a verification code is generated and sent
// to the user's email, and the code is stored for validation.
// @Summary Authenticate user and send login code
// @Description This endpoint verifies user credentials by checking the email and password. If valid, it sends a login code to the user's email for further verification. The request must be a POST method with a JSON body containing "email" and "password". The login code is stored temporarily for verification purposes.
// @Tags User Authentication
// @Accept json
// @Produce json
// @Param credentials body map[string]string true "Email and Password"
// @Success 200 {object} string "Login code sent to email"
// @Failure 400 {string} string "Invalid Input: Error parsing JSON body"
// @Failure 401 {string} string "Unauthorized: Invalid email or password"
// @Failure 500 {string} string "Internal Server Error: Failed to send login code"
// @Router /api/user/login [post]
func Login(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
		return
	}

	var data map[string]string

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user models.User

	if err := database.DB.Where("email = ?", data["email"]).First(&user).Error; err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data["password"])); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	code := generateCode()
	if err := sendLoginCode(user.Email, code); err != nil {
		http.Error(w, "Failed to send login code", http.StatusInternalServerError)
		return
	}

	loginCodeStore.Lock()
	loginCodeStore.data[user.Email] = code
	loginCodeStore.Unlock()

	response := map[string]interface{}{
		"status":  http.StatusOK,
		"message": "Login code sent to email",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyLoginCode handles the verification of login codes sent to users.
// It checks the provided login code against the stored code for the given email,
// and if valid, generates a JWT token and sets it as a cookie in the response.
// This function expects a POST request with a JSON body containing the email and login code.
// @Summary Verify the login code and generate JWT token
// @Description This endpoint verifies the provided login code for the specified email. If the code is valid, a JWT token is generated and set as a cookie in the response.
// @Tags User Authentication
// @Accept json
// @Produce json
// @Param body body map[string]string true "Email and login code"
// @Success 200 {object} string "Login successful"
// @Failure 400 {string} string "Invalid Input: Error parsing request body"
// @Failure 401 {string} string "Unauthorized: Invalid or expired login code"
// @Failure 500 {string} string "Internal Server Error: Failed to generate token"
// @Router /api/user/verify-login-code [post]
func VerifyLoginCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data map[string]string
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	loginCodeStore.RLock()
	storedCode, exists := loginCodeStore.data[data["email"]]
	loginCodeStore.RUnlock()

	if !exists || storedCode != data["code"] {
		http.Error(w, "Invalid or expired login code", http.StatusUnauthorized)
		return
	}

	loginCodeStore.Lock()
	delete(loginCodeStore.data, data["email"])
	loginCodeStore.Unlock()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    data["email"],
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte("your-256-bit-secret"))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	response := map[string]interface{}{
		"status":  http.StatusOK,
		"message": "Login successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Println("User logged in successfully.")
}

// LogOut handles the user sign-out process by clearing the JWT token cookie.
// It expects a POST request and will invalidate the user's session by removing the JWT token cookie.
// @Summary Sign out the user and clear the JWT token
// @Description This endpoint logs out the user by clearing the JWT token cookie. The request must be a POST method. Upon successful sign-out, the JWT token is removed from the cookies, and a success message is returned.
// @Tags User Authentication
// @Accept json
// @Produce json
// @Success 200 {object} string "Successfully signed out"
// @Failure 405 {string} string "Method Not Allowed: Invalid request method"
// @Router /api/user/logout [post]
func LogOut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    "",
		Expires:  time.Now().Add(-24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	response := map[string]interface{}{
		"status":  http.StatusOK,
		"message": "Successfully signed out",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
