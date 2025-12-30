package types

import (
	"CVWO-Backend/models"
	"time"
)

type PostResponse struct {
	ID        uint       `json:"id"`
	TopicID   uint       `json:"topicId"`
	UserID    uint       `json:"userId"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	EditedAt  *time.Time `json:"editedAt,omitempty"`
	Author    UserPublic `json:"author"`
}

func ToPostResponse(p models.Post) PostResponse {
	return PostResponse{
		ID:        p.ID,
		TopicID:   p.TopicID,
		UserID:    p.UserID,
		Title:     p.Title,
		Body:      p.Body,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		EditedAt:  p.EditedAt,
		Author:    ToUserPublic(p.User),
	}
}
