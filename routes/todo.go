package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
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
		activityGroupID := fmt.Sprintf("%v", m["activity_group_id"])
		activityID, err := strconv.Atoi(activityGroupID)
		if err != nil {
			return parser.GetResponseNoData(c, 400, "Bad Request", err.Error())
		}
		todo.ActivityID = activityID
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
	stmt, err := db.Prepare("INSERT INTO todos(created_at, updated_at, activity_group_id, title, is_active, priority) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}
	defer stmt.Close()

	resp, err := stmt.Exec(todo.CreatedAt, todo.UpdatedAt, todo.ActivityID, todo.Title, todo.IsActive, todo.Priority)
	if err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}

	newMap := make(map[string]interface{})
	for _, key := range []string{"created_at", "updated_at", "id", "title", "activity_group_id", "is_active",
		"deleted_at", "priority"} {
		newMap[key] = m[key]
	}

	newMap["id"], err = resp.LastInsertId()
	if err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
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

	stmt, err := db.Prepare("SELECT id, created_at, updated_at, deleted_at, activity_group_id, title, is_active, priority FROM todos where deleted_at IS NULL AND id = ?")
	if err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}
	defer stmt.Close()

	rows, err := stmt.Query(c.Params("id"))
	if err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&todo.ID, &todo.CreatedAt, &todo.UpdatedAt, &todo.DeletedAt, &todo.ActivityID, &todo.Title,
			&todo.IsActive, &todo.Priority); err != nil {
			return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
		}
	}

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
	var wg sync.WaitGroup
	var errs []error

	activityGroupID = string(c.Request().URI().QueryArgs().Peek("activity_group_id"))

	key := []byte("todos-items")
	var firstID, lastID int

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

	wg.Add(1)
	go func() {
		defer wg.Done()
		stmt, err := db.Prepare("SELECT id FROM todos WHERE deleted_at IS NULL ORDER BY id ASC LIMIT 1")
		if err != nil {
			errs = append(errs, err)
		}
		defer stmt.Close()

		rows, err := stmt.Query()
		if err != nil {
			errs = append(errs, err)
		}
		defer rows.Close()

		for rows.Next() {
			rows.Scan(&firstID)
		}
		errs = append(errs, nil)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		stmt, err := db.Prepare("SELECT id FROM todos WHERE deleted_at IS NULL ORDER BY id DESC LIMIT 1")
		if err != nil {
			errs = append(errs, err)
		}
		defer stmt.Close()

		rows, err := stmt.Query()
		if err != nil {
			errs = append(errs, err)
		}
		defer rows.Close()

		for rows.Next() {
			rows.Scan(&lastID)
		}
		errs = append(errs, nil)
	}()

	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
		}
	}

	bucketSize := 300
	resultCount := 0

	resultChannel := make(chan []*models.TodoModel)
	for beginID := firstID; beginID <= lastID; beginID += bucketSize {
		endID := beginID + bucketSize
		go func(beginID, endID int) {
			var tempTodos []*models.TodoModel
			if activityGroupID != "" {
				stmt, err := db.Prepare("SELECT id, created_at, updated_at, deleted_at, activity_group_id, title, is_active, priority FROM todos WHERE deleted_at IS NUll and id >= ? AND id < ? AND activity_group_id = ?")
				if err != nil {
					errs = append(errs, err)
				}
				defer stmt.Close()

				rows, err := stmt.Query(beginID, endID, activityGroupID)
				if err != nil {
					errs = append(errs, err)
				}
				defer rows.Close()

				for rows.Next() {
					var tempTodo models.TodoModel
					if err := rows.Scan(&tempTodo.ID, &tempTodo.CreatedAt, &tempTodo.UpdatedAt, &tempTodo.DeletedAt,
						&tempTodo.ActivityID, &tempTodo.Title, &tempTodo.IsActive, &tempTodo.Priority); err != nil {
						errs = append(errs, err)
					}
					tempTodos = append(tempTodos, &tempTodo)
				}
			} else {
				stmt, err := db.Prepare("SELECT id, created_at, updated_at, deleted_at, activity_group_id, title, is_active, priority FROM todos WHERE deleted_at IS NUll and id >= ? AND id < ?")
				if err != nil {
					errs = append(errs, err)
				}
				defer stmt.Close()

				rows, err := stmt.Query(beginID, endID)
				if err != nil {
					errs = append(errs, err)
				}
				defer rows.Close()

				for rows.Next() {
					var tempTodo models.TodoModel
					if err := rows.Scan(&tempTodo.ID, &tempTodo.CreatedAt, &tempTodo.UpdatedAt, &tempTodo.DeletedAt,
						&tempTodo.ActivityID, &tempTodo.Title, &tempTodo.IsActive, &tempTodo.Priority); err != nil {
						errs = append(errs, err)
					}
					tempTodos = append(tempTodos, &tempTodo)
				}
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

// func deleteTodo(c *fiber.Ctx) error {
// 	db := database.DBConn
// 	resp := db.Model(&models.TodoModel{}).Where("id = ? and deleted_at is null", c.Params("id")).
// 		Update("deleted_at", time.Now().UTC())

// 	if resp.Error != nil || resp.RowsAffected == 0 {
// 		return parser.GetResponseNoData(c, 404, "Not Found", fmt.Sprintf("Todo with ID %s Not Found",
// 			c.Params("id")))
// 	}

// 	cache := cache.Cache
// 	cache.Del([]byte(fmt.Sprintf("todo-%s", c.Params("id"))))
// 	return parser.GetResponseNoData(c, 200, "Success", "Success")
// }

// func updateTodo(c *fiber.Ctx) error {
// 	db := database.DBConn
// 	tempTodo := new(models.TodoModel)

// 	m := make(map[string]interface{})
// 	if err := c.BodyParser(&m); err != nil {
// 		return parser.GetResponseNoData(c, 400, "Bad Request", err.Error())
// 	}

// 	if m["title"] != nil {
// 		switch m["title"].(type) {
// 		case string:
// 			tempTodo.Title = m["title"].(string)
// 		}
// 	}

// 	switch m["is_active"].(type) {
// 	case string:
// 		tempTodo.IsActive = m["is_active"].(string)
// 	case bool:
// 		tempTodo.IsActive = "1"
// 	case nil:
// 		tempTodo.IsActive = ""
// 	default:
// 		tempTodo.IsActive = fmt.Sprintf("%v", m["is_active"])
// 	}

// 	switch m["priority"].(type) {
// 	case nil:
// 		tempTodo.Priority = ""
// 	default:
// 		tempTodo.Priority = fmt.Sprintf("%v", m["priority"])
// 	}

// 	var todo *models.TodoModel
// 	db.Where("deleted_at is null").Find(&todo, c.Params("id"))
// 	if todo.ID == 0 {
// 		return parser.GetResponseNoData(c, 404, "Not Found",
// 			fmt.Sprintf("Todo with ID %s Not Found", c.Params("id")))
// 	}

// 	if tempTodo.Title != "" {
// 		todo.Title = tempTodo.Title
// 	}

// 	if tempTodo.IsActive != "" {
// 		todo.IsActive = tempTodo.IsActive
// 	}

// 	if tempTodo.Priority != "" {
// 		todo.Priority = tempTodo.Priority
// 	}

// 	todo.UpdatedAt = time.Now().UTC()

// 	isValid, errMessage := todo.Validate()
// 	if !isValid {
// 		return parser.GetResponseNoData(c, 400, "Bad Request", errMessage)
// 	}
// 	db.Save(&todo)

// 	cache := cache.Cache
// 	cache.Del([]byte(fmt.Sprintf("todo-%s", c.Params("id"))))
// 	return parser.GetTodoResponse(c, 200, "Success", "Success", todo)
// }
