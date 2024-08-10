package models

type User struct {
	Id       string `json:"id" gorm:"primaryKey"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"-"`
}
