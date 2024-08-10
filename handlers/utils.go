package handlers

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/karan-singh-17/Quick-Mail/database"
	"github.com/karan-singh-17/Quick-Mail/models"
)

var loginCodeStore = struct {
	sync.RWMutex
	data map[string]string
}{data: make(map[string]string)}

var tempStore = struct {
	sync.Mutex
	data map[string]models.User
}{data: make(map[string]models.User)}

func generateCode() string {
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}

func GenerateID(email string) string {

	hash := sha256.Sum256([]byte(email))

	encoded := base32.StdEncoding.EncodeToString(hash[:])

	encoded = strings.TrimRight(encoded, "=")
	encoded = strings.ToLower(encoded)

	if len(encoded) > 10 {
		return encoded[:10]
	}

	return encoded
}

func generateToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GetUser(w http.ResponseWriter, r *http.Request) (models.User, error) {
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return models.User{}, err
	}

	tokenString := cookie.Value
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-256-bit-secret"), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return models.User{}, err
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return models.User{}, err
	}

	var user models.User
	if err := database.DB.Where("email = ?", claims.Issuer).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return models.User{}, err
	}

	return user, err
}

func isValidEmail(email string) bool {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func fetchRecipientsFromCSV(csvLink string) ([]string, error) {
	var recipients []string

	if strings.Contains(csvLink, "docs.google.com/spreadsheets/") {
		csvLink = convertGoogleSheetToCSV(csvLink)
	}

	// Download the CSV file
	resp, err := http.Get(csvLink)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download csv file")
	}
	reader := csv.NewReader(resp.Body)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) > 0 {
			recipients = append(recipients, strings.TrimSpace(record[0]))
		}
	}

	return recipients, nil
}
func convertGoogleSheetToCSV(link string) string {
	// Extract the sheet ID from the Google Sheets link
	re := regexp.MustCompile(`https://docs.google.com/spreadsheets/d/([a-zA-Z0-9-_]+)`)
	match := re.FindStringSubmatch(link)

	if len(match) > 1 {
		sheetID := match[1]
		return fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/export?format=csv", sheetID)
	}

	return link // Return the original link if it doesn't match the pattern
}
func extractRecipientsFromFilePath(filePath string) ([]string, error) {
	var recipients []string

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Assuming the CSV contains emails in the first column
		if len(record) > 0 {
			recipients = append(recipients, strings.TrimSpace(record[0]))
		}
	}

	return recipients, nil
}
func validateSingleFilledField(fields ...string) bool {
	filledCount := 0
	for _, field := range fields {
		if strings.TrimSpace(field) != "" {
			filledCount++
		}
	}
	return filledCount == 1
}

func fetchHTMLFromLink(link string) (string, error) {
	resp, err := http.Get(link)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error fetching HTML content, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func readHTMLFromFilePath(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
