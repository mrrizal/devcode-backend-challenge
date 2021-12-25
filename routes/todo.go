package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mrrizal/devcode-backend-challenge/database"
	"github.com/mrrizal/devcode-backend-challenge/models"
	"github.com/mrrizal/devcode-backend-challenge/parser"
)

func createTodo(c *fiber.Ctx) error {
	db := database.DBConn

	todo := new(models.TodoModel)
	if err := c.BodyParser(todo); err != nil {
		return parser.GetResponseNoData(c, 400, "Bad Request", "title cannot be null")
	}

	isValid, message := todo.Validate()
	if !isValid {
		return parser.GetResponseNoData(c, 400, "Bad Request", message)
	}

	now := time.Now().UTC()
	todo.CreatedAt = now
	todo.UpdatedAt = now
	todo.IsActive = "1"
	todo.Priority = "very-high"

	type deletedAtStruct struct {
		DeletedAt *time.Time
	}
	deletedAt := deletedAtStruct{}
	db.Model(&models.ActivityModel{}).Select("deleted_at").Where("id = ?", todo.ActivityID).Find(&deletedAt)
	if deletedAt.DeletedAt != nil {
		return parser.GetResponseNoData(c, 400, "Bad Request", "acvity group has been deleted")
	}

	result := db.Create(&todo)
	if result.Error != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", result.Error.Error())
	}
	return parser.GetTodoResponse(c, 201, "Success", "Success", todo)
}
