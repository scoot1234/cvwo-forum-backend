package models

import "time"

type User struct {
	ID uint `gorm:"primaryKey" json:"id"`
	Username string `gorm:"size:32;not null;uniqueIndex" json:"username"`
	PasswordHash string `gorm:"not null" json:"-"`
	Role string `gorm:"size:16;not null;default:user" json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Posts    []Post    `json:"-"` 
	Comments []Comment `json:"-"`
}
