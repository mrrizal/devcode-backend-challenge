package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

var PriorityType [5]string = [5]string{"very-high", "high", "normal", "low", "very-low"}

type TodoModel struct {
	gorm.Model
	ID            int           `gorm:"primaryKey"`
	ActivityID    int           `json:"activity_group_id"`
	ActivityModel ActivityModel `gorm:"foreignKey:ActivityID;references:ID"`
	Title         string        `json:"title"`
	IsActive      string        `json:"is_active"`
	Priority      string        `json:"priority"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time `gorm:"index:todo_deleted_at"`
}

func (TodoModel) TableName() string {
	return "todo"
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (todo TodoModel) Validate() (bool, string) {
	if todo.Title == "" {
		return false, "title cannot be null"
	}

	if todo.ActivityID == 0 {
		return false, "activity_group_id cannot be null"
	}

	todo.Priority = strings.ToLower(todo.Priority)
	if !stringInSlice(todo.Priority, PriorityType[:]) {
		return false, "Data truncated for column 'priority' at row 1"
	}

	return true, ""
}
