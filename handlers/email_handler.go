package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"sync"

	"github.com/karan-singh-17/Quick-Mail/database"
	"github.com/karan-singh-17/Quick-Mail/models"
)

var from = os.Getenv("from")
var password = os.Getenv("password")
var smtpHost = os.Getenv("smtpHost")
var smtpPort = os.Getenv("smtpPort")

func sendVerificationEmail(email, token string) error {

	to := email
	auth := smtp.PlainAuth("", from, password, smtpHost)

	htmlContent, err := os.ReadFile("templates/verify_mail.html")
	if err != nil {
		return fmt.Errorf("error reading HTML template: %v", err)
	}

	subject := "Subject: Verify Your Account\n"

	htmlText := strings.ReplaceAll(string(htmlContent), "{{TOKEN}}", token)

	message := []byte(subject + "MIME-Version: 1.0\nContent-Type: text/html; charset=\"UTF-8\";\n\n" + htmlText)

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	return smtp.SendMail(addr, auth, from, []string{to}, message)
}

func sendLoginCode(email, code string) error {
	to := email

	auth := smtp.PlainAuth("", from, password, smtpHost)

	htmlContent, err := os.ReadFile("templates/send_mail_temp.html")
	if err != nil {
		return fmt.Errorf("error reading HTML template: %v", err)
	}

	subject := "Subject: Your Login Code\n"

	htmlText := strings.ReplaceAll(string(htmlContent), "{{MESSAGE}}", code)

	message := []byte(subject + "MIME-Version: 1.0\nContent-Type: text/html; charset=\"UTF-8\";\n\n" + htmlText)

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	return smtp.SendMail(addr, auth, from, []string{to}, message)
}

// Execute Group
// @Summary execute/run the group
// @Description starts the process of sending emails to the recipients. Make sure you are logged in and are the owner of the group.
// @Tags Groups
// @Accept json
// @Produce json
// @Param group_id body map[string]string true "Group ID" example({"group_id": "example-group-id"})
// @Success 200
// @Failure 400 {object} string "Invalid Input"
// @Failure 401 {object} string "Unauthorized"
// @Failure 404 {object} string "Group not found"
// @Failure 500 {object} string "Error in sending mails"
// @Router /api/group/execute-group [post]
// @Example { "group_id": "example-group-id" }
func SendMailToGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
		return
	}

	var g_id map[string]string
	if err := json.NewDecoder(r.Body).Decode(&g_id); err != nil {
		http.Error(w, "Invalid Input", http.StatusBadRequest)
		return
	}

	var curr_grp models.Group

	if err := database.DB.Where("group_id = ?", g_id["group_id"]).First(&curr_grp).Error; err != nil {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	curr_user, err := GetUser(w, r)

	if err != nil {
		panic(err)
	}

	if curr_grp.Owner_ID != curr_user.Id {
		http.Error(w, "You are not the owner of this group. Access Denied", http.StatusUnauthorized)
		return
	}

	errdf := sendmailtogrp(curr_grp)

	if errdf != nil {
		http.Error(w, "Error in sending mails", http.StatusInternalServerError)
		return
	}

}

func sendmailtogrp(group models.Group) error {
	auth := smtp.PlainAuth("", from, password, smtpHost)
	recipients := strings.Split(group.Recipients, ",")

	validRecipients := []string{}
	for _, recipient := range recipients {
		recipient = strings.TrimSpace(recipient)
		if recipient != "" && isValidEmail(recipient) {
			validRecipients = append(validRecipients, recipient)
		}
	}

	if len(validRecipients) == 0 {
		return fmt.Errorf("no valid recipients found")
	}

	subject := "Subject: " + group.Subject + "\n"
	htmlContent, err := os.ReadFile("templates/send_mail_temp.html")
	if err != nil {
		return fmt.Errorf("error reading HTML template: %v", err)
	}

	htmlText := strings.ReplaceAll(string(htmlContent), "{{MESSAGE}}", group.Message)
	message := []byte(subject + "MIME-Version: 1.0\nContent-Type: text/html; charset=\"UTF-8\";\n\n" + htmlText)
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	var wg sync.WaitGroup
	errCh := make(chan error, len(validRecipients))

	for _, recipient := range validRecipients {
		wg.Add(1)
		go func(recipient string) {
			defer wg.Done()
			err := smtp.SendMail(addr, auth, from, []string{recipient}, message)
			if err != nil {
				errCh <- fmt.Errorf("error sending email to %s: %v", recipient, err)
			}
		}(recipient)
	}

	wg.Wait()
	close(errCh)

	if len(errCh) > 0 {
		return <-errCh
	}
	return nil
}
