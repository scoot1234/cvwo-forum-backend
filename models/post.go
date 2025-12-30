package models

import "time"

type Post struct {
	ID uint `gorm:"primaryKey" json:"id"`
	TopicID uint `gorm:"not null;index" json:"topicId"`
	UserID  uint `gorm:"not null;index" json:"userId"`
	Title string `gorm:"size:120;not null" json:"title"`
	Body  string `gorm:"type:text;not null" json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Topic Topic `gorm:"foreignKey:TopicID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	EditedAt *time.Time `gorm:"index" json:"editedAt,omitempty"`
	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Comments []Comment `json:"-"`
}
