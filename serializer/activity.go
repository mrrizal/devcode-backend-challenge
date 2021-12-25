package serializer

type BaseResponse struct {
	Message string `json:"message"`
}

type Activity struct {
	ID        int     `json:"id"`
	Email     string  `json:"email"`
	Title     string  `json:"title"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	DeletedAt *string `json:"deleted_at"`
}

type ActivityResponse struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Data    *Activity `json:"data"`
}

type ResponseNoData struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Data    map[string]string `json:"data"`
}

type ActivitiesResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    *[]Activity `json:"data"`
}
