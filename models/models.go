package models

import (
	"time"
)

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"uniqueIndex"`
	Password string
}

type JournalEntry struct {
	ID        uint `gorm:"primaryKey"`
	Content   string
	Timestamp time.Time
	UserID    uint
}

type ExerciseEntry struct {
	ID        uint `gorm:"primaryKey"`
	Activity  string
	Duration  float64 // in minutes
	Timestamp time.Time
	UserID    uint
}

type NutritionEntry struct {
	ID        uint   `gorm:"primaryKey"`
	MealType  string // e.g., Breakfast, Lunch, etc.
	Content   string
	Timestamp time.Time
	UserID    uint
}

type Reminder struct {
	ID      uint `gorm:"primaryKey"`
	Content string
	Time    time.Time
	UserID  uint
}
