package controllers

import (
	"errors"
	"net/http"
	"strings"

	"CVWO-Backend/models"
	"CVWO-Backend/types"
	"CVWO-Backend/utils"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type PostsController struct {
	DB *gorm.DB
}

func NewPostsController(db *gorm.DB) *PostsController {
	return &PostsController{DB: db}
}

func (c *PostsController) RegisterRoutes(r chi.Router) {
	r.Get("/topics/{topicId}/posts", c.GetPostsByTopic)
	r.Post("/topics/{topicId}/posts", c.CreatePost)
	
	r.Get("/posts/{postId}", c.GetPostByID)
	r.Patch("/posts/{postId}", c.UpdatePost)
	r.Delete("/posts/{postId}", c.DeletePost)
}

func (c *PostsController) GetPostsByTopic(w http.ResponseWriter, r *http.Request) {
	topicID, err := utils.ParseUintParam(r, "topicId")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid topicId")
		return
	}

	var topic models.Topic
	if err := c.DB.First(&topic, topicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteError(w, http.StatusNotFound, "topic not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch topic")
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))

	dbq := c.DB.
		Where("topic_id = ?", topicID).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username")
		}).
		Order("created_at DESC")

	if q != "" {
		like := "%" + q + "%"
		dbq = dbq.Where("(title ILIKE ? OR body ILIKE ?)", like, like)
	}

	var posts []models.Post
	if err := dbq.Find(&posts).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch posts")
		return
	}

	out := make([]types.PostResponse, 0, len(posts))
	for _, p := range posts {
		out = append(out, types.ToPostResponse(p))
	}

	utils.WriteJSON(w, http.StatusOK, out)
}

func (c *PostsController) CreatePost(w http.ResponseWriter, r *http.Request) {
	topicID, err := utils.ParseUintParam(r, "topicId")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid topicId")
		return
	}

	var topic models.Topic
	if err := c.DB.First(&topic, topicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteError(w, http.StatusNotFound, "topic not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "db error checking topic")
		return
	}

	type createPostRequest struct {
		UserID uint   `json:"userId"`
		Title  string `json:"title"`
		Body   string `json:"body"`
	}
	var req createPostRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)

	if req.UserID == 0 {
		utils.WriteError(w, http.StatusBadRequest, "userId is required")
		return
	}
	if req.Title == "" {
		utils.WriteError(w, http.StatusBadRequest, "title cannot be empty")
		return
	}
	if len(req.Title) > 120 {
		utils.WriteError(w, http.StatusBadRequest, "title too long (max 120)")
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

	post := models.Post{
		TopicID: topicID,
		UserID:  req.UserID,
		Title:   req.Title,
		Body:    req.Body,
	}

	if err := c.DB.Create(&post).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to create post")
		return
	}

	post.User = user

	utils.WriteJSON(w, http.StatusCreated, types.ToPostResponse(post))
}

func (c *PostsController) UpdatePost(w http.ResponseWriter, r *http.Request) {
	postID, err := utils.ParseUintParam(r, "postId")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid postId")
		return
	}

	type updatePostRequest struct {
		UserID uint    `json:"userId"`
		Title  *string `json:"title,omitempty"`
		Body   *string `json:"body,omitempty"`
	}
	var req updatePostRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.UserID == 0 {
		utils.WriteError(w, http.StatusBadRequest, "userId is required")
		return
	}
	if req.Title == nil && req.Body == nil {
		utils.WriteError(w, http.StatusBadRequest, "nothing to update")
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

	var requester models.User
	if err := c.DB.Select("id", "role").First(&requester, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteError(w, http.StatusBadRequest, "user not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "db error checking user")
		return
	}

	owner := post.UserID == requester.ID
	if !(isPrivileged(requester) || owner) {
		utils.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	updates := map[string]any{}
	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if t == "" {
			utils.WriteError(w, http.StatusBadRequest, "title cannot be empty")
			return
		}
		if len(t) > 120 {
			utils.WriteError(w, http.StatusBadRequest, "title too long (max 120)")
			return
		}
		updates["title"] = t
	}
	if req.Body != nil {
		b := strings.TrimSpace(*req.Body)
		if b == "" {
			utils.WriteError(w, http.StatusBadRequest, "body cannot be empty")
			return
		}
		updates["body"] = b
	}

	if err := c.DB.Model(&post).Updates(updates).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to update post")
		return
	}

	var updated models.Post
	if err := c.DB.
		Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username") }).
		First(&updated, postID).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch updated post")
		return
	}

	utils.WriteJSON(w, http.StatusOK, types.ToPostResponse(updated))
}

func (c *PostsController) DeletePost(w http.ResponseWriter, r *http.Request) {
	postID, err := utils.ParseUintParam(r, "postId")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid postId")
		return
	}

	type deletePostRequest struct {
		UserID uint `json:"userId"`
	}
	var req deletePostRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.UserID == 0 {
		utils.WriteError(w, http.StatusBadRequest, "userId is required")
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

	var requester models.User
	if err := c.DB.Select("id", "role").First(&requester, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteError(w, http.StatusBadRequest, "user not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "db error checking user")
		return
	}

	owner := post.UserID == requester.ID
	if !(isPrivileged(requester) || owner) {
		utils.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := c.DB.Delete(&post).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to delete post")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *PostsController) GetPostByID(w http.ResponseWriter, r *http.Request) {
	postID, err := utils.ParseUintParam(r, "postId")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid postId")
		return
	}

	var post models.Post
	if err := c.DB.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username")
		}).
		First(&post, postID).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteError(w, http.StatusNotFound, "post not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch post")
		return
	}

	utils.WriteJSON(w, http.StatusOK, types.ToPostResponse(post))
}
