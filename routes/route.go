package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mrrizal/devcode-backend-challenge/database"
	"github.com/mrrizal/devcode-backend-challenge/models"
	"github.com/mrrizal/devcode-backend-challenge/serializer"
	"github.com/mrrizal/devcode-backend-challenge/utils"
)

func root(c *fiber.Ctx) error {
	resp := serializer.BaseResponse{
		Message: "Welcome to API TODO",
	}
	return c.JSON(resp)
}

func createActivity(c *fiber.Ctx) error {
	db := database.DBConn

	activity := new(models.ActivityModel)
	if err := c.BodyParser(activity); err != nil {
		return utils.GetActivityErrorResponse(c, 400, "Bad Request", "title cannot be null")

	}

	isValid, message := activity.Validate()
	if !isValid {
		return utils.GetActivityErrorResponse(c, 400, "Bad Request", message)
	}

	now := time.Now()
	activity.CreatedAt = now
	activity.UpdatedAt = now

	result := db.Create(&activity)
	if result.Error != nil {
		return utils.GetActivityErrorResponse(c, 500, "Internal Server Error", result.Error.Error())
	}
	return utils.GetActivityResponse(c, 201, "Success", "Success", activity)
}

func SetupRoutes(app *fiber.App) {
	app.Get("/", root)
	app.Post("/activity-groups/", createActivity)
}
