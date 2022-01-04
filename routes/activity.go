package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mrrizal/devcode-backend-challenge/cache"
	"github.com/mrrizal/devcode-backend-challenge/database"
	"github.com/mrrizal/devcode-backend-challenge/models"
	"github.com/mrrizal/devcode-backend-challenge/parser"
)

func createActivity(c *fiber.Ctx) error {
	db := database.DBConn

	activity := new(models.ActivityModel)
	if err := c.BodyParser(activity); err != nil {
		return parser.GetResponseNoData(c, 400, "Bad Request", "title cannot be null")
	}

	isValid, message := activity.Validate()
	if !isValid {
		return parser.GetResponseNoData(c, 400, "Bad Request", message)
	}

	now := time.Now()
	activity.CreatedAt = now
	activity.UpdatedAt = now

	stmt, err := db.Prepare("INSERT INTO activities (email, title, created_at, updated_at) VALUES (?, ?, ?, ?)")
	if err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}
	defer stmt.Close()

	resp, err := stmt.Exec(activity.Email, activity.Title, activity.CreatedAt, activity.UpdatedAt)
	if err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}

	insertedID, err := resp.LastInsertId()
	if err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}

	activity.ID = int(insertedID)

	return parser.GetActivityResponse(c, 201, "Success", "Success", activity)
}

func getActivity(c *fiber.Ctx) error {
	db := database.DBConn
	activity := new(models.ActivityModel)
	cache := cache.Cache
	expire := 120
	key := []byte(fmt.Sprintf("activity-%s", c.Params("id")))

	got, err := cache.Get(key)
	if err == nil {
		if err := json.Unmarshal(got, &activity); err != nil {
			return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
		}
		return parser.GetActivityResponse(c, 200, "Success", "Success", activity)
	}

	stmt, err := db.Prepare("SELECT id, email, title, created_at, updated_at, deleted_at FROM activities WHERE id = ? AND deleted_at IS NULL LIMIT 1")
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
		if err := rows.Scan(&activity.ID, &activity.Email, &activity.Title, &activity.CreatedAt, &activity.UpdatedAt,
			&activity.DeletedAt); err != nil {
			return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
		}
	}

	if err := rows.Err(); err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}

	if activity.ID == 0 {
		return parser.GetResponseNoData(c, 404, "Not Found",
			fmt.Sprintf("Activity with ID %s Not Found", c.Params("id")))
	}

	// set cache
	activitiesBytes := new(bytes.Buffer)
	json.NewEncoder(activitiesBytes).Encode(activity)
	if err := cache.Set(key, activitiesBytes.Bytes(), expire); err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}

	return parser.GetActivityResponse(c, 200, "Success", "Success", activity)
}

func getActivities(c *fiber.Ctx) error {
	db := database.DBConn
	cache := cache.Cache
	var activities []*models.ActivityModel
	expire := 120
	key := []byte("activities")
	var firstID, lastID int
	var wg sync.WaitGroup
	var errs []error

	// get data from cache
	got, err := cache.Get(key)
	if err == nil {
		if err := json.Unmarshal(got, &activities); err != nil {
			return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
		}
		return parser.GetActivitiesResponse(c, 200, "Success", "Success", activities)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		stmt, err := db.Prepare("SELECT id FROM activities WHERE deleted_at IS NULL ORDER BY id ASC LIMIT 1")
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
		stmt, err := db.Prepare("SELECT id FROM activities WHERE deleted_at IS NULL ORDER BY id DESC LIMIT 1")
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

	resultChannel := make(chan []*models.ActivityModel)
	errs = []error{}
	for beginID := firstID; beginID <= lastID; beginID += bucketSize {
		endID := beginID + bucketSize
		go func(beginID, endID int) {
			var tempActivities []*models.ActivityModel
			stmt, err := db.Prepare("SELECT id, email, title, created_at, updated_at, deleted_at FROM activities WHERE deleted_at IS NULL AND id >= ? AND id < ?")
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
				var tempActivity models.ActivityModel
				err := rows.Scan(&tempActivity.ID, &tempActivity.Email, &tempActivity.Title, &tempActivity.CreatedAt,
					&tempActivity.UpdatedAt, &tempActivity.DeletedAt)
				if err != nil {
					errs = append(errs, err)
				} else {
					tempActivities = append(tempActivities, &tempActivity)
				}
			}
			resultChannel <- tempActivities
		}(beginID, endID)
		resultCount += 1
	}

	for i := 0; i < resultCount; i++ {
		tempActivities := <-resultChannel
		activities = append(activities, tempActivities...)
	}
	sort.Slice(activities, func(i, j int) bool { return activities[i].ID < activities[j].ID })

	// set cache
	activitiesBytes := new(bytes.Buffer)
	json.NewEncoder(activitiesBytes).Encode(activities)
	if err := cache.Set(key, activitiesBytes.Bytes(), expire); err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}

	return parser.GetActivitiesResponse(c, 200, "Success", "Success", activities)
}

func updateActivity(c *fiber.Ctx) error {
	db := database.DBConn
	title := struct {
		Title string `json:"title"`
	}{}

	if err := c.BodyParser(&title); err != nil {
		return parser.GetResponseNoData(c, 400, "Bad Request", "title cannot be null")
	}

	if title.Title == "" {
		return parser.GetResponseNoData(c, 400, "Bad Request", "title cannot be null")
	}

	// var activity *models.ActivityModel
	stmt, err := db.Prepare("UPDATE activities SET title=?, updated_at=? WHERE deleted_at IS NULL and id=? LIMIT 1")
	if err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}
	defer stmt.Close()

	resp, err := stmt.Exec(title.Title, time.Now().UTC(), c.Params("id"))
	if err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}

	count, err := resp.RowsAffected()
	if err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}

	if count == 0 {
		return parser.GetResponseNoData(c, 404, "Not Found",
			fmt.Sprintf("Activity with ID %s Not Found", c.Params("id")))
	}

	stmt, err = db.Prepare("SELECT id, email, title, created_at, updated_at, deleted_at FROM activities WHERE id = ? AND deleted_at IS NULL LIMIT 1")
	if err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}
	defer stmt.Close()

	rows, err := stmt.Query(c.Params("id"))
	if err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}
	defer rows.Close()

	var activity models.ActivityModel
	for rows.Next() {
		if err := rows.Scan(&activity.ID, &activity.Email, &activity.Title, &activity.CreatedAt, &activity.UpdatedAt,
			&activity.DeletedAt); err != nil {
			return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
		}
	}

	cache := cache.Cache
	cache.Del([]byte(fmt.Sprintf("activity-%s", c.Params("id"))))
	return parser.GetActivityResponse(c, 200, "Success", "Success", &activity)
}

func deleteActivity(c *fiber.Ctx) error {
	db := database.DBConn
	// var activity *models.ActivityModel
	stmt, err := db.Prepare("UPDATE activities SET deleted_at=? WHERE deleted_at IS NULL and id=? LIMIT 1")
	if err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}
	defer stmt.Close()

	resp, err := stmt.Exec(time.Now().UTC(), c.Params("id"))
	if err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}

	count, err := resp.RowsAffected()
	if err != nil {
		return parser.GetResponseNoData(c, 500, "Internal Server Error", err.Error())
	}

	if count == 0 {
		return parser.GetResponseNoData(c, 404, "Not Found",
			fmt.Sprintf("Activity with ID %s Not Found", c.Params("id")))
	}

	cache := cache.Cache
	cache.Del([]byte(fmt.Sprintf("activity-%s", c.Params("id"))))
	return parser.GetResponseNoData(c, 200, "Success", "Success")
}
