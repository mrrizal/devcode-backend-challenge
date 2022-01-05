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

type CreateTodoResponse struct {
	Status  string                  `json:"status"`
	Message string                  `json:"message"`
	Data    *map[string]interface{} `json:"data"`
}

type CreateTodoBodyRequest struct {
	ID              interface{} `json:"id"`
	ActivityGroupID interface{} `json:"activity_group_id"`
	Title           interface{} `json:"title"`
	IsActive        interface{} `json:"is_active"`
	Priority        interface{} `json:"priority"`
	CreatedAt       interface{} `json:"created_at"`
	UpdatedAt       interface{} `json:"updated_at"`
	DeletedAt       interface{} `json:"deleted_at"`
}
