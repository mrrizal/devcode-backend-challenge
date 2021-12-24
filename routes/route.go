package routes

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mrrizal/devcode-backend-challenge/database"
	"github.com/mrrizal/devcode-backend-challenge/models"
	"github.com/mrrizal/devcode-backend-challenge/parser"
	"github.com/mrrizal/devcode-backend-challenge/serializer"
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
		return parser.GetActivityErrorResponse(c, 400, "Bad Request", "title cannot be null")

	}

	isValid, message := activity.Validate()
	if !isValid {
		return parser.GetActivityErrorResponse(c, 400, "Bad Request", message)
	}

	now := time.Now()
	activity.CreatedAt = now
	activity.UpdatedAt = now

	result := db.Create(&activity)
	if result.Error != nil {
		return parser.GetActivityErrorResponse(c, 500, "Internal Server Error", result.Error.Error())
	}
	return parser.GetActivityResponse(c, 201, "Success", "Success", activity)
}

func getActivity(c *fiber.Ctx) error {
	db := database.DBConn
	activity := new(models.ActivityModel)
	db.Find(&activity, c.Params("id"))
	if activity.ID == 0 {
		return parser.GetActivityErrorResponse(c, 404, "Not Found",
			fmt.Sprintf("Activity with ID %s Not Found", c.Params("id")))
	}
	return parser.GetActivityResponse(c, 200, "Success", "Success", activity)
}

func getActivities(c *fiber.Ctx) error {
	db := database.DBConn
	var activities []*models.ActivityModel
	// todo: insert 10k - 50k data, then try to optimize
	db.Where("deleted_at is null").Find(&activities)
	return parser.GetActivitiesResponse(c, 200, "Success", "Success", activities)
}

func SetupRoutes(app *fiber.App) {
	app.Get("/", root)
	app.Post("/activity-groups/", createActivity)
	app.Get("/activity-groups/", getActivities)
	app.Get("/activity-groups/:id", getActivity)
}
