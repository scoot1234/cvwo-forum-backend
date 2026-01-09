package controllers

import (
	"errors"
	"net/http"
	"strings"

	"CVWO-Backend/models"
	"CVWO-Backend/utils"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthController struct {
	DB *gorm.DB
}

func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{DB: db}
}

func (c *AuthController) RegisterRoutes(r chi.Router) {
	r.Post("/auth/signup", c.SignUp)
	r.Post("/auth/login", c.Login)
}

func (c *AuthController) SignUp(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		utils.WriteError(w, http.StatusBadRequest, "username cannot be empty")
		return
	}
	if len(req.Username) > 32 {
		utils.WriteError(w, http.StatusBadRequest, "username too long (max 32)")
		return
	}
	if len(req.Password) < 8 {
		utils.WriteError(w, http.StatusBadRequest, "password too short (min 8)")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user := models.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Role:         "user",
	}

	if err := c.DB.Create(&user).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			utils.WriteError(w, http.StatusConflict, "username already taken")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{
		"message": "signup ok",
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		utils.WriteError(w, http.StatusBadRequest, "username and password required")
		return
	}

	var user models.User
	if err := c.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "failed to query user")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "login ok",
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}
