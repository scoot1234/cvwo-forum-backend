package db

import (
	"CVWO-Backend/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SeedDefaultTopics(gdb *gorm.DB) error {
	defaultTopics := []models.Topic{
		{Title: "Tech", Description: "Gadgets, programming, AI, and tech news."},
		{Title: "Games", Description: "Gaming discussion: PC/console/mobile, esports, and releases."},
		{Title: "Lifestyle", Description: "Daily life, productivity, wellness, and habits."},
		{Title: "Music", Description: "Artists, genres, recommendations, and live shows."},
		{Title: "Automotive", Description: "Cars, bikes, mods, reviews, and maintenance."},
		{Title: "Culture", Description: "Pop culture, trends, media, and society."},
	}

	return gdb.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "title"}},
			DoNothing: true,
		}).
		Create(&defaultTopics).Error
}
