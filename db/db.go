package db

import (
	"go-health/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitializeDB() (*gorm.DB, error) {
	Db, err := gorm.Open(sqlite.Open("health.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	Db.AutoMigrate(&models.User{}, &models.JournalEntry{}, &models.ExerciseEntry{}, &models.NutritionEntry{}, &models.Reminder{})

	return Db, nil
}
