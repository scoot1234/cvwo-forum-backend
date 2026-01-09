package models

import "time"

type Comment struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	PostID uint `gorm:"not null;index" json:"postId"`
	UserID uint `gorm:"not null;index" json:"userId"`

	Body string `gorm:"type:text;not null" json:"body"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	EditedAt  *time.Time `gorm:"index" json:"editedAt,omitempty"`

	Post Post `gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}
