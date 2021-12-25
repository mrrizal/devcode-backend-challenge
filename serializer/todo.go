package serializer

type Todo struct {
	ID              int     `json:"id"`
	ActivityGroupID string  `json:"activity_group_id"`
	Title           string  `json:"title"`
	IsActive        string  `json:"is_active"`
	Priority        string  `json:"priority"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	DeletedAt       *string `json:"deleted_at"`
}

type TodoResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    *Todo  `json:"data"`
}

type TodosResponse struct {
	Status  string  `json:"status"`
	Message string  `json:"message"`
	Data    *[]Todo `json:"data"`
}
