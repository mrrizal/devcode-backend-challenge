package models

import "time"

type Activity struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	Email     string    `json:"email"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `gorm:"index" json:"deleted_at"`
}

func (activity *Activity) Validate() bool {
	// todo: email must be valid email regex
	return activity.Title != ""
}

type ActivityResponse struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Data    *Activity `json:"data"`
}

type ActivityErrorResponse struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Data    map[string]string `json:"data"`
}

type ActivitiesResponse struct {
	Status  string     `json:"status"`
	Message string     `json:"message"`
	Data    []Activity `json:"data"`
}
