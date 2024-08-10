package models

type Group struct {
	Group_ID   string `gorm:"primaryKey" json:"group_id"`
	Name       string `json:"name"`
	Owner_ID   string `json:"owner_id"`
	Recipients string `json:"recipients"`
	Subject    string `json:"subject"`
	Message    string `json:"message"`
}
