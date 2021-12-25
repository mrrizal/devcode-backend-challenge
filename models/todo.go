package models

import (
	"time"

	"gorm.io/gorm"
)

type TodoModel struct {
	gorm.Model
	ID            int           `gorm:"primaryKey"`
	ActivityID    int           `json:"activity_group_id"`
	ActivityModel ActivityModel `gorm:"foreignKey:ActivityID;references:ID"`
	Title         string        `json:"title"`
	IsActive      string
	Priority      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time `gorm:"index:todo_deleted_at"`
}

func (TodoModel) TableName() string {
	return "todo"
}

func (todo TodoModel) Validate() (bool, string) {
	if todo.Title == "" {
		return false, "title cannot be null"
	}

	if todo.ActivityID == 0 {
		return false, "activity_group_id cannot be null"
	}

	return true, ""
}
