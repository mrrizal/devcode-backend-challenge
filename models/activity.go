package models

import (
	"net/mail"
	"time"
)

type ActivityModel struct {
	ID        int
	Email     string
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (ActivityModel) TableName() string {
	return "activities"
}

func (activity *ActivityModel) Validate() (bool, string) {
	if activity.Email != "" {
		_, err := mail.ParseAddress(activity.Email)
		if err != nil {
			return false, "invalid email address"
		}
	}

	if activity.Title != "" {
		return true, ""
	}
	return false, "title cannot be null"
}
