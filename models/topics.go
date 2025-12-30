package models

import "time"

type Topic struct {
	ID uint `gorm:"primaryKey" json:"id"`
	Title string `gorm:"size:100;not null;uniqueIndex" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	CreatedByUserID *uint `gorm:"index" json:"createdByUserId,omitempty"`
	CreatedByUser   *User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Posts []Post `json:"-"`
}

	