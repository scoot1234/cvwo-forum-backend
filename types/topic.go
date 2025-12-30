package types

import (
	"CVWO-Backend/models"
	"time"
)

type TopicResponse struct {
	ID              uint       `json:"id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	CreatedByUserID *uint      `json:"createdByUserId,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	Author          *UserPublic `json:"author,omitempty"`
}

func ToTopicResponse(t models.Topic) TopicResponse {
	var author *UserPublic
	if t.CreatedByUserID != nil {
		u := ToUserPublic(*t.CreatedByUser)
		author = &u
	}
	return TopicResponse{
		ID:              t.ID,
		Title:           t.Title,
		Description:     t.Description,
		CreatedByUserID: t.CreatedByUserID,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
		Author:          author,
	}
}
