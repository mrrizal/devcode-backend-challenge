package parser

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mrrizal/devcode-backend-challenge/models"
	"github.com/mrrizal/devcode-backend-challenge/serializer"
)

func parseActivity(activityModel *models.ActivityModel) serializer.Activity {
	activity := serializer.Activity{
		ID:        activityModel.ID,
		Email:     activityModel.Email,
		Title:     activityModel.Title,
		CreatedAt: fmt.Sprintf("%sZ", activityModel.CreatedAt.Format("2006-01-02T15:04:05.000")),
		UpdatedAt: fmt.Sprintf("%sZ", activityModel.UpdatedAt.Format("2006-01-02T15:04:05.000")),
	}
	if activityModel.DeletedAt != nil {
		deletedAt := *activityModel.DeletedAt
		*activity.DeletedAt = fmt.Sprintf("%sZ", deletedAt.Format("2006-01-02T15:04:05.000"))
	}
	return activity
}

func parseTodo(todoModel *models.TodoModel) serializer.Todo {
	todo := serializer.Todo{
		ID:              todoModel.ID,
		ActivityGroupID: strconv.Itoa(todoModel.ActivityID),
		Title:           todoModel.Title,
		IsActive:        todoModel.IsActive,
		Priority:        todoModel.Priority,
		CreatedAt:       fmt.Sprintf("%sZ", todoModel.CreatedAt.Format("2006-01-02T15:04:05.000")),
		UpdatedAt:       fmt.Sprintf("%sZ", todoModel.UpdatedAt.Format("2006-01-02T15:04:05.000")),
	}
	if todoModel.DeletedAt != nil {
		deletedAt := *todoModel.DeletedAt
		*todo.DeletedAt = fmt.Sprintf("%sZ", deletedAt.Format("2006-01-02T15:04:05.000"))
	}
	return todo
}

func GetResponseNoData(c *fiber.Ctx, statusCode int, status, message string) error {
	resp := serializer.ResponseNoData{
		Status:  status,
		Message: message,
		Data:    make(map[string]string),
	}
	return json.NewEncoder(c.Type("json", "utf-8").Status(statusCode).Response().BodyWriter()).Encode(resp)

}

func GetActivityResponse(c *fiber.Ctx, statusCode int, status, message string, activityModel *models.ActivityModel) error {
	activity := parseActivity(activityModel)
	resp := serializer.ActivityResponse{
		Status:  status,
		Message: message,
		Data:    &activity,
	}
	return json.NewEncoder(c.Type("json", "utf-8").Status(statusCode).Response().BodyWriter()).Encode(resp)
}

func GetActivitiesResponse(c *fiber.Ctx, statusCode int, status, message string, activitiesModel []*models.ActivityModel) error {
	var activities []serializer.Activity
	for _, activityModel := range activitiesModel {
		activity := parseActivity(activityModel)
		activities = append(activities, activity)
	}

	resp := serializer.ActivitiesResponse{
		Status:  status,
		Message: message,
		Data:    &activities,
	}
	return json.NewEncoder(c.Type("json", "utf-8").Status(statusCode).Response().BodyWriter()).Encode(resp)
}

func GetTodoResponse(c *fiber.Ctx, statusCode int, status, message string, todoModel *models.TodoModel) error {
	todo := parseTodo(todoModel)
	resp := serializer.TodoResponse{
		Status:  status,
		Message: message,
		Data:    &todo,
	}
	return json.NewEncoder(c.Type("json", "utf-8").Status(statusCode).Response().BodyWriter()).Encode(resp)
}

func TodoCreateResponse(c *fiber.Ctx, statusCode int, status, message string, m *map[string]interface{}) error {
	resp := serializer.CreateTodoResponse{
		Status:  status,
		Message: message,
		Data:    m,
	}
	return json.NewEncoder(c.Type("json", "utf-8").Status(statusCode).Response().BodyWriter()).Encode(resp)
}

func GetTodosResponse(c *fiber.Ctx, statusCode int, status, message string, todosModel []*models.TodoModel) error {
	var todos []serializer.Todo
	for _, todoModel := range todosModel {
		todo := parseTodo(todoModel)
		todos = append(todos, todo)
	}

	resp := serializer.TodosResponse{
		Status:  status,
		Message: message,
		Data:    &todos,
	}

	temp := make([]serializer.Todo, 0)
	if len(todos) == 0 {
		resp.Data = &temp
	}
	return json.NewEncoder(c.Type("json", "utf-8").Status(statusCode).Response().BodyWriter()).Encode(resp)
}
