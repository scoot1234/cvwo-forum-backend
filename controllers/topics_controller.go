package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"CVWO-Backend/models"
	"CVWO-Backend/types"
	"CVWO-Backend/utils"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type TopicsController struct {
	DB *gorm.DB
}

func NewTopicsController(db *gorm.DB) *TopicsController {
	return &TopicsController{DB: db}
}

func (c *TopicsController) RegisterRoutes(r chi.Router) {
	r.Get("/topics", c.GetTopics)
	r.Post("/topics", c.CreateTopic)

	r.Patch("/topics/{topicId}", c.UpdateTopic)
	r.Delete("/topics/{topicId}", c.DeleteTopic)
}

func (c *TopicsController) GetTopics(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	dbq := c.DB.
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username")
		}).
		Order("created_at DESC")

	if q != "" {
		like := "%" + q + "%"
		dbq = dbq.Where("title ILIKE ?", like)
	}

	var topics []models.Topic
	if err := dbq.Find(&topics).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch topics")
		return
	}

	out := make([]types.TopicResponse, 0, len(topics))
	for _, t := range topics {
		out = append(out, types.ToTopicResponse(t))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}


func (c *TopicsController) CreateTopic(w http.ResponseWriter, r *http.Request) {
	type createTopicRequest struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		UserID      *uint  `json:"userId,omitempty"`
	}

	var req createTopicRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)

	if req.Title == "" {
		utils.WriteError(w, http.StatusBadRequest, "title cannot be empty")
		return
	}
	if len(req.Title) > 100 {
		utils.WriteError(w, http.StatusBadRequest, "title too long (max 100)")
		return
	}
	if len(req.Description) > 500 {
		utils.WriteError(w, http.StatusBadRequest, "description too long (max 500)")
		return
	}

	var author models.User
	if req.UserID != nil {
		if err := c.DB.Select("id", "username", "role").First(&author, *req.UserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				utils.WriteError(w, http.StatusBadRequest, "user not found")
				return
			}
			utils.WriteError(w, http.StatusInternalServerError, "db error checking user")
			return
		}
	}

	topic := models.Topic{
		Title:           req.Title,
		Description:     req.Description,
		CreatedByUserID: req.UserID,
	}

	if err := c.DB.Create(&topic).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			utils.WriteError(w, http.StatusConflict, "topic title already exists")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "failed to create topic")
		return
	}

	if req.UserID != nil {
		topic.CreatedByUser = &author
	}

	utils.WriteJSON(w, http.StatusCreated, types.ToTopicResponse(topic))
}

func (c *TopicsController) UpdateTopic(w http.ResponseWriter, r *http.Request) {
	topicID, err := utils.ParseUintParam(r, "topicId")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid topicId")
		return
	}

	type updateTopicRequest struct {
		UserID      uint    `json:"userId"`
		Title       *string `json:"title,omitempty"`
		Description *string `json:"description,omitempty"`
	}

	var req updateTopicRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.UserID == 0 {
		utils.WriteError(w, http.StatusBadRequest, "userId is required")
		return
	}
	if req.Title == nil && req.Description == nil {
		utils.WriteError(w, http.StatusBadRequest, "nothing to update")
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

	var requester models.User
	if err := c.DB.Select("id", "role").First(&requester, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteError(w, http.StatusBadRequest, "user not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "db error checking user")
		return
	}

	isOwner := topic.CreatedByUserID != nil && *topic.CreatedByUserID == requester.ID
	if !(isPrivileged(requester) || isOwner) {
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
		if len(t) > 100 {
			utils.WriteError(w, http.StatusBadRequest, "title too long (max 100)")
			return
		}
		updates["title"] = t
	}

	if req.Description != nil {
		d := strings.TrimSpace(*req.Description)
		if len(d) > 500 {
			utils.WriteError(w, http.StatusBadRequest, "description too long (max 500)")
			return
		}
		updates["description"] = d
	}

	if err := c.DB.Model(&topic).Updates(updates).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			utils.WriteError(w, http.StatusConflict, "topic title already exists")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "failed to update topic")
		return
	}

	var updated models.Topic
	if err := c.DB.
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username") }).
		First(&updated, topicID).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch updated topic")
		return
	}

	utils.WriteJSON(w, http.StatusOK, types.ToTopicResponse(updated))
}

func (c *TopicsController) DeleteTopic(w http.ResponseWriter, r *http.Request) {
	topicID, err := utils.ParseUintParam(r, "topicId")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid topicId")
		return
	}

	type deleteTopicRequest struct {
		UserID uint `json:"userId"`
	}

	var req deleteTopicRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.UserID == 0 {
		utils.WriteError(w, http.StatusBadRequest, "userId is required")
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

	var requester models.User
	if err := c.DB.Select("id", "role").First(&requester, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteError(w, http.StatusBadRequest, "user not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "db error checking user")
		return
	}

	isOwner := topic.CreatedByUserID != nil && *topic.CreatedByUserID == requester.ID
	if !(isPrivileged(requester) || isOwner) {
		utils.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := c.DB.Delete(&topic).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to delete topic")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
