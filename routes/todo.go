package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mrrizal/devcode-backend-challenge/cache"
	"github.com/mrrizal/devcode-backend-challenge/database"
	"github.com/mrrizal/devcode-backend-challenge/models"
	"github.com/mrrizal/devcode-backend-challenge/parser"
)

func createTodo(c *fiber.Ctx) error {
	db := database.DBConn
	todo := new(models.TodoModel)

	m := make(map[string]interface{})
	if err := c.BodyParser(&m); err != nil {
		return parser.GetResponseNoData(c, 400, "Bad Request", err.Error())
	}

	if m["activity_group_id"] != nil && m["activity_group_id"] != "" {
		type deletedAtStruct struct {
			ID        int
			DeletedAt *time.Time
		}
		activityGroupID := fmt.Sprintf("%v", m["activity_group_id"])
		deletedAt := deletedAtStruct{}
		db.Model(&models.ActivityModel{}).Select("deleted_at", "id").Where("id = ?", activityGroupID).Find(&deletedAt)
		if deletedAt.DeletedAt != nil || deletedAt.ID == 0 {
			return parser.GetResponseNoData(c, 404, "Not Found",
				fmt.Sprintf("Activity with activity_group_id %v Not Found", activityGroupID))
		}
		todo.ActivityID = deletedAt.ID
	}

	if m["title"] != nil {
		switch m["title"].(type) {
		case string:
			todo.Title = m["title"].(string)
		}
	}

	switch m["is_active"].(type) {
	case string:
		todo.IsActive = m["is_active"].(string)
	case bool:
		todo.IsActive = "1"
		m["is_active"] = true
	case nil:
		todo.IsActive = "1"
		m["is_active"] = true
	default:
		todo.IsActive = fmt.Sprintf("%v", m["is_active"])
	}

	switch m["priority"].(type) {
	case nil:
		todo.Priority = "very-high"
		m["priority"] = "very-high"
	default:
		todo.Priority = fmt.Sprintf("%v", m["priority"])
	}

	now := time.Now().UTC()
	todo.CreatedAt = now
	todo.UpdatedAt = now
	m["created_at"] = fmt.Sprintf("%sZ", now.Format("2006-01-02T15:04:05.000"))
	m["updated_at"] = fmt.Sprintf("%sZ", now.Format("2006-01-02T15:04:05.000"))
	m["deleted_at"] = nil

	isValid, message := todo.Validate()
	if !isValid {
		return parser.GetResponseNoData(c, 400, "Bad Request", message)
	}

	todo.Priority = strings.ToLower(todo.Priority)
	result := db.Create(&todo)
	newMap := make(map[string]interface{})
	for _, key := range []string{"created_at", "updated_at", "id", "title", "activity_group_id", "is_active",
		"deleted_at", "priority"} {
		newMap[key] = m[key]
	}
	newMap["id"] = todo.ID

	if result.Error != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", result.Error.Error())
	}
	return parser.TodoCreateResponse(c, 201, "Success", "Success", &newMap)
}

func getTodo(c *fiber.Ctx) error {
	db := database.DBConn
	todo := new(models.TodoModel)
	cache := cache.Cache
	expire := 120
	key := []byte(fmt.Sprintf("todo-%s", c.Params("id")))

	got, err := cache.Get(key)
	if err == nil {
		if err := json.Unmarshal(got, &todo); err != nil {
			return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
		}
		return parser.GetTodoResponse(c, 200, "Success", "Success", todo)
	}

	db.Where("deleted_at is null").Find(&todo, c.Params("id"))
	if todo.ID == 0 {
		return parser.GetResponseNoData(c, 404, "Not Found",
			fmt.Sprintf("Todo with ID %s Not Found", c.Params("id")))
	}

	// set cache
	activitiesBytes := new(bytes.Buffer)
	json.NewEncoder(activitiesBytes).Encode(todo)
	if err := cache.Set(key, activitiesBytes.Bytes(), expire); err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}

	return parser.GetTodoResponse(c, 200, "Success", "Success", todo)
}

func getTodos(c *fiber.Ctx) error {
	var activityGroupID string
	db := database.DBConn
	cache := cache.Cache
	var todos []*models.TodoModel
	expire := 120

	activityGroupID = string(c.Request().URI().QueryArgs().Peek("activity_group_id"))

	key := []byte("todos-items")
	var firstID, lastID struct {
		ID int
	}

	if activityGroupID != "" {
		key = []byte(fmt.Sprintf("%s-activity_group_id-%s", string(key), activityGroupID))
	}

	// get data from cache
	got, err := cache.Get(key)
	if err == nil {
		if err := json.Unmarshal(got, &todos); err != nil {
			return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
		}
		return parser.GetTodosResponse(c, 200, "Success", "Success", todos)
	}

	bucketSize := 500
	resultCount := 0
	db.Model(&models.TodoModel{}).First(&firstID)
	db.Model(&models.TodoModel{}).Last(&lastID)

	resultChannel := make(chan []*models.TodoModel)
	for beginID := firstID.ID; beginID <= lastID.ID; beginID += bucketSize {
		endID := beginID + bucketSize
		go func(beginID, endID int) {
			var tempTodos []*models.TodoModel
			if activityGroupID != "" {
				db.Where("deleted_at is null and id >= ? and id < ? and activity_id = ?",
					beginID, endID, activityGroupID).Find(&tempTodos)
			} else {
				db.Where("deleted_at is null and id >= ? and id < ?", beginID, endID).Find(&tempTodos)
			}
			resultChannel <- tempTodos
		}(beginID, endID)
		resultCount += 1
	}

	for i := 0; i < resultCount; i++ {
		tempTodos := <-resultChannel
		todos = append(todos, tempTodos...)
	}
	sort.Slice(todos, func(i, j int) bool { return todos[i].ID < todos[j].ID })

	// set cache
	todosBytes := new(bytes.Buffer)
	json.NewEncoder(todosBytes).Encode(todos)
	if err := cache.Set(key, todosBytes.Bytes(), expire); err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}

	return parser.GetTodosResponse(c, 200, "Success", "Success", todos)
}

func deleteTodo(c *fiber.Ctx) error {
	db := database.DBConn
	resp := db.Model(&models.TodoModel{}).Where("id = ? and deleted_at is null", c.Params("id")).
		Update("deleted_at", time.Now().UTC())

	if resp.Error != nil || resp.RowsAffected == 0 {
		return parser.GetResponseNoData(c, 404, "Not Found", fmt.Sprintf("Todo with ID %s Not Found",
			c.Params("id")))
	}

	cache := cache.Cache
	cache.Del([]byte(fmt.Sprintf("todo-%s", c.Params("id"))))
	return parser.GetResponseNoData(c, 200, "Success", "Success")
}

func updateTodo(c *fiber.Ctx) error {
	db := database.DBConn
	tempTodo := new(models.TodoModel)

	m := make(map[string]interface{})
	if err := c.BodyParser(&m); err != nil {
		return parser.GetResponseNoData(c, 400, "Bad Request", err.Error())
	}

	if m["title"] != nil {
		switch m["title"].(type) {
		case string:
			tempTodo.Title = m["title"].(string)
		}
	}

	switch m["is_active"].(type) {
	case string:
		tempTodo.IsActive = m["is_active"].(string)
	case bool:
		tempTodo.IsActive = "1"
	case nil:
		tempTodo.IsActive = ""
	default:
		tempTodo.IsActive = fmt.Sprintf("%v", m["is_active"])
	}

	switch m["priority"].(type) {
	case nil:
		tempTodo.Priority = ""
	default:
		tempTodo.Priority = fmt.Sprintf("%v", m["priority"])
	}

	var todo *models.TodoModel
	db.Where("deleted_at is null").Find(&todo, c.Params("id"))
	if todo.ID == 0 {
		return parser.GetResponseNoData(c, 404, "Not Found",
			fmt.Sprintf("Todo with ID %s Not Found", c.Params("id")))
	}

	if tempTodo.Title != "" {
		todo.Title = tempTodo.Title
	}

	if tempTodo.IsActive != "" {
		todo.IsActive = tempTodo.IsActive
	}

	if tempTodo.Priority != "" {
		todo.Priority = tempTodo.Priority
	}

	todo.UpdatedAt = time.Now().UTC()

	isValid, errMessage := todo.Validate()
	if !isValid {
		return parser.GetResponseNoData(c, 400, "Bad Request", errMessage)
	}
	db.Save(&todo)

	cache := cache.Cache
	cache.Del([]byte(fmt.Sprintf("todo-%s", c.Params("id"))))
	return parser.GetTodoResponse(c, 200, "Success", "Success", todo)
}
