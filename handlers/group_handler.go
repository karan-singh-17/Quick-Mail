package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/karan-singh-17/Quick-Mail/database"
	"github.com/karan-singh-17/Quick-Mail/models"
)

type GroupData struct {
	Name         string   `json:"name"`
	Recipients   []string `json:"recipients"`
	CSVLink      string   `json:"csv_link,omitempty"`
	Subject      string   `json:"subject"`
	Message      string   `json:"message"`
	CSVFilePath  string   `json:"csv_file_path,omitempty"`
	HTMLLink     string   `json:"html_link,omitempty"`
	HTMLFilePath string   `json:"html_path,omitempty"`
}

// Post Groups
// @Summary creates a new group
// @Description creates a new group. Make sure you are logged in and follow the parameter rules.
// @Tags Groups
// @Accept json
// @Produce json
// @Param Group body GroupData true "Group"
// @Success 201 {object} map[string]interface{} "Successful response with group details"
// @Failure 400 {object} string "Bad Request"
// @Failure 401 {object} string "Unauthorized"
// @Failure 404 {object} string "Not Found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /api/group/create-group [post]
// @security jwt_token
func CreateGroup(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
		return
	}

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

	var data GroupData

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Error Parsing Data", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(data.Name) == "" || (strings.TrimSpace(data.Message) == "" && strings.TrimSpace(data.HTMLFilePath) == "" && strings.TrimSpace(data.HTMLLink) == "") || (data.Recipients == nil && strings.TrimSpace(data.CSVLink) == "" && strings.TrimSpace(data.CSVFilePath) == "") || strings.TrimSpace(data.Subject) == "" {
		http.Error(w, "Wrong Inputs... Please refer the docs", http.StatusBadRequest)
		return
	}

	if !validateSingleFilledField(data.Message, data.HTMLFilePath, data.HTMLLink) {
		http.Error(w, "Only one of Message, HTMLFilePath, or HTMLLink can be provided", http.StatusBadRequest)
		return
	}

	if !validateSingleFilledField(data.CSVFilePath, data.CSVLink) {
		http.Error(w, "Only one of CSVFilePath or CSVLink can be provided", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(data.CSVFilePath) != "" {
		fileRecipients, err := extractRecipientsFromFilePath(data.CSVFilePath)
		if err != nil {
			http.Error(w, "Error reading CSV file from path", http.StatusInternalServerError)
			return
		}
		data.Recipients = append(data.Recipients, fileRecipients...)
	}

	if strings.TrimSpace(data.CSVLink) != "" {
		csvRecipients, err := fetchRecipientsFromCSV(data.CSVLink)
		if err != nil {
			http.Error(w, "Error fetching recipients from CSV", http.StatusInternalServerError)
			return
		}
		data.Recipients = append(data.Recipients, csvRecipients...)
	}

	if strings.TrimSpace(data.HTMLLink) != "" {
		htmlContent, err := fetchHTMLFromLink(data.HTMLLink)
		if err != nil {
			http.Error(w, "Error fetching HTML from link", http.StatusInternalServerError)
			return
		}
		data.Message = htmlContent
	}

	if strings.TrimSpace(data.HTMLFilePath) != "" {
		htmlContent, err := readHTMLFromFilePath(data.HTMLFilePath)
		if err != nil {
			http.Error(w, "Error reading HTML from file path", http.StatusInternalServerError)
			return
		}
		data.Message = htmlContent
	}
	tokenid, err := generateToken()
	if err != nil {
		panic(err)
	}

	group := models.Group{
		Group_ID:   "g-" + tokenid,
		Name:       data.Name,
		Owner_ID:   user.Id,
		Recipients: strings.Join(data.Recipients, ","),
		Subject:    data.Subject,
		Message:    data.Message,
	}

	if err := database.DB.Create(&group).Error; err != nil {
		http.Error(w, "Error creating group", http.StatusInternalServerError)
		return
	}
	response := map[string]interface{}{
		"status":  http.StatusCreated,
		"message": "Group created",
		"group":   group,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Println("Group created:", group)
}

// Get Groups
// @Summary return all groups of the current user
// @Description returns all the groups created by the user. Make sure you are logged in and have a valid jwt_token
// @Tags Groups
// @Success 200 {object} []models.Group
// @Router /api/group/get-groups [get]
// @security jwt_token
func GetAllGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
		return
	}

	curr_user, err := GetUser(w, r)

	if err != nil {
		http.Error(w, "Unauthorized or wrong ", http.StatusUnauthorized)
		return
	}

	var groups []models.Group
	if err := database.DB.Where("owner_id = ?", curr_user.Id).Find(&groups).Error; err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Respond with the list of groups as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(groups); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Edit Group
// @Summary edit an existing group
// @Description edits the details of a group. Make sure you are logged in and are the owner of the group.
// @Tags Groups
// @Accept json
// @Produce json
// @Param Group body models.Group true "Group"
// @Success 200 {object} map[string]string "message"
// @Router /api/group/edit-group [put]
// @security jwt_token
func EditGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed!!", http.StatusMethodNotAllowed)
		return
	}

	var data map[string]string

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var grp models.Group
	if err := database.DB.Where("group_id = ?", data["group_id"]).First(&grp).Error; err != nil {
		http.Error(w, "Group Not Found", http.StatusNotFound)
		return
	}

	curr_user, err := GetUser(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if grp.Owner_ID != curr_user.Id {
		http.Error(w, "You are not the owner of this group. Access Denied", http.StatusUnauthorized)
		return
	}

	if name, ok := data["name"]; ok {
		grp.Name = name
	}
	if recipients, ok := data["recipients"]; ok {
		grp.Recipients = recipients
	}
	if subject, ok := data["subject"]; ok {
		grp.Subject = subject
	}
	if message, ok := data["message"]; ok {
		grp.Message = message
	}

	if err := database.DB.Save(&grp).Error; err != nil {
		http.Error(w, "Error updating group", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Group updated successfully"})
}

// Delete Group
// @Summary delete a group
// @Description deletes an existing group. Make sure you are logged in and are the owner of the group.
// @Tags Groups
// @Accept json
// @Produce json
// @Param group_id body map[string]string true "Group ID" example(`{"group_id": "example-group-id"}`)
// @Success 200 {object} map[string]string "message"
// @Router /api/group/delete-group [delete]
// @security jwt_token
func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var data map[string]string

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Error Parsing Data", http.StatusBadRequest)
		return
	}

	var grp models.Group
	if err := database.DB.Where("group_id = ?", data["group_id"]).First(&grp).Error; err != nil {
		http.Error(w, "Group Not Found", http.StatusNotFound)
		return
	}

	curr_user, err := GetUser(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if grp.Owner_ID != curr_user.Id {
		http.Error(w, "You are not the owner of this group. Access Denied", http.StatusUnauthorized)
		return
	}

	if err := database.DB.Delete(grp).Error; err != nil {
		http.Error(w, "Unable to delete group", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Group deleted successfully"})
}
