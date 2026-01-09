package controllers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"CVWO-Backend/models"
	"CVWO-Backend/types"
	"CVWO-Backend/utils"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type CommentsController struct {
	DB *gorm.DB
}

func NewCommentsController(db *gorm.DB) *CommentsController {
	return &CommentsController{DB: db}
}

func (c *CommentsController) RegisterRoutes(r chi.Router) {
	r.Get("/posts/{postId}/comments", c.GetCommentsByPost)
	r.Post("/posts/{postId}/comments", c.CreateComment)

	r.Patch("/comments/{commentId}", c.UpdateComment)
	r.Delete("/comments/{commentId}", c.DeleteComment)
}

func (c *CommentsController) GetCommentsByPost(w http.ResponseWriter, r *http.Request) {
	postID, err := utils.ParseUintParam(r, "postId")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid postId")
		return
	}

	var post models.Post
	if err := c.DB.First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteError(w, http.StatusNotFound, "post not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch post")
		return
	}

	var comments []models.Comment
	if err := c.DB.
		Where("post_id = ?", postID).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username")
		}).
		Order("created_at ASC").
		Find(&comments).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch comments")
		return
	}

	out := make([]types.CommentResponse, 0, len(comments))
	for _, cm := range comments {
		out = append(out, types.ToCommentResponse(cm))
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (c *CommentsController) CreateComment(w http.ResponseWriter, r *http.Request) {
	postID, err := utils.ParseUintParam(r, "postId")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid postId")
		return
	}

	var post models.Post
	if err := c.DB.First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteError(w, http.StatusNotFound, "post not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "db error checking post")
		return
	}

	type createCommentRequest struct {
		UserID uint   `json:"userId"`
		Body   string `json:"body"`
	}

	var req createCommentRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	req.Body = strings.TrimSpace(req.Body)

	if req.UserID == 0 {
		utils.WriteError(w, http.StatusBadRequest, "userId is required")
		return
	}
	if req.Body == "" {
		utils.WriteError(w, http.StatusBadRequest, "body cannot be empty")
		return
	}

	var user models.User
	if err := c.DB.Select("id", "username", "role").First(&user, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteError(w, http.StatusBadRequest, "user not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "db error checking user")
		return
	}

	comment := models.Comment{
		PostID: postID,
		UserID: req.UserID,
		Body:   req.Body,
	}

	if err := c.DB.Create(&comment).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to create comment")
		return
	}

	comment.User = user
	utils.WriteJSON(w, http.StatusCreated, types.ToCommentResponse(comment))
}

func (c *CommentsController) UpdateComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := utils.ParseUintParam(r, "commentId")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid commentId")
		return
	}

	type updateCommentRequest struct {
		UserID uint    `json:"userId"`
		Body   *string `json:"body,omitempty"`
	}

	var req updateCommentRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if req.UserID == 0 {
		utils.WriteError(w, http.StatusBadRequest, "userId is required")
		return
	}
	if req.Body == nil {
		utils.WriteError(w, http.StatusBadRequest, "nothing to update")
		return
	}

	var comment models.Comment
	if err := c.DB.First(&comment, commentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteError(w, http.StatusNotFound, "comment not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch comment")
		return
	}

	var requester models.User
	if err := c.DB.Select("id", "role").First(&requester, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteError(w, http.StatusBadRequest, "user not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "db error checking user")
		return
	}

	owner := comment.UserID == requester.ID
	if !(isPrivileged(requester) || owner) {
		utils.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	body := strings.TrimSpace(*req.Body)
	if body == "" {
		utils.WriteError(w, http.StatusBadRequest, "body cannot be empty")
		return
	}

	now := time.Now()
	if err := c.DB.Model(&comment).Updates(map[string]any{
		"body":      body,
		"edited_at": &now,
	}).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to update comment")
		return
	}

	var updated models.Comment
	if err := c.DB.
		Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username") }).
		First(&updated, commentID).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch updated comment")
		return
	}

	utils.WriteJSON(w, http.StatusOK, types.ToCommentResponse(updated))
}

func (c *CommentsController) DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := utils.ParseUintParam(r, "commentId")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid commentId")
		return
	}

	type deleteCommentRequest struct {
		UserID uint `json:"userId"`
	}

	var req deleteCommentRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if req.UserID == 0 {
		utils.WriteError(w, http.StatusBadRequest, "userId is required")
		return
	}

	var comment models.Comment
	if err := c.DB.First(&comment, commentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteError(w, http.StatusNotFound, "comment not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch comment")
		return
	}

	var requester models.User
	if err := c.DB.Select("id", "role").First(&requester, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteError(w, http.StatusBadRequest, "user not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "db error checking user")
		return
	}

	owner := comment.UserID == requester.ID
	if !(isPrivileged(requester) || owner) {
		utils.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := c.DB.Delete(&comment).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to delete comment")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
