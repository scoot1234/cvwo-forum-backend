package types

import "CVWO-Backend/models"

type UserPublic struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

func ToUserPublic(u models.User) UserPublic {
	return UserPublic{ID: u.ID, Username: u.Username}
}
