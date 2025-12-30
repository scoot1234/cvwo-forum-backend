package controllers

import "CVWO-Backend/models"

func isPrivileged(u models.User) bool {
	return u.Role == "admin" || u.Role == "moderator"
}
