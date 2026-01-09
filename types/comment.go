package types

import (
	"time"

	"CVWO-Backend/models"
)

type CommentResponse struct {
	ID     uint `json:"id"`
	PostID uint `json:"postId"`
	UserID uint `json:"userId"`
	Body   string `json:"body"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	EditedAt  *time.Time `json:"editedAt,omitempty"`

	Author UserPublic `json:"author"`
}

func ToCommentResponse(c models.Comment) CommentResponse {
	return CommentResponse{
		ID:        c.ID,
		PostID:    c.PostID,
		UserID:    c.UserID,
		Body:      c.Body,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		EditedAt:  c.EditedAt,
		Author:    ToUserPublic(c.User),
	}
}
