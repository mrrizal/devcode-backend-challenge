package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mrrizal/devcode-backend-challenge/serializer"
)

func root(c *fiber.Ctx) error {
	resp := serializer.BaseResponse{
		Message: "Welcome to API TODO",
	}
	return c.JSON(resp)
}

func SetupRoutes(app *fiber.App) {
	app.Get("/", root)
	app.Post("/activity-groups/", createActivity)
	app.Get("/activity-groups/", getActivities)
	app.Get("/activity-groups/:id", getActivity)
	app.Patch("/activity-groups/:id", updateActivity)
	app.Delete("/activity-groups/:id", deleteActivity)

	app.Post("/todo-items/", createTodo)
	app.Get("/todo-items/:id", getTodo)
	app.Get("/todo-items/", getTodos)
	app.Delete("/todo-items/:id", deleteTodo)
	app.Patch("/todo-items/:id", updateTodo)
}
