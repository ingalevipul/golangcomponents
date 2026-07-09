package model

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Name      string `gorm:"type:varchar(100);not null"`
	Email     string `gorm:"not null"`
	Password  string `gorm:"not null"`
	Status    string `gorm:"column:user_status;default:active"`
	Age       int    `gorm:"not null;check:age>18"`
	Posts     []Post
	Languages []Language `gorm:"many2many:language_users"`
}
type Post struct {
	gorm.Model
	PostID uint
	Title  string
	UserID uint
	User   User
}

type Language struct {
	gorm.Model
	Name string
}
type UnamePass struct {
	Username string `gorm:"primaryKey;not null"`
	Password string `gorm:"not null"`
}

func (u *UnamePass) TableName() string {
	return "unamepass"
}
