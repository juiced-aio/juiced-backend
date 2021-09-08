package api

import (
	"backend.juicedbot.io/juiced.api/controller"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var app *fiber.App

// StartServer launches the local server that hosts the API for communication between the app and the backend
func StartServer() {
	app = fiber.New(fiber.Config{
		StrictRouting: true,
		CaseSensitive: true,
		Concurrency:   1024 * 1024 * 16,
		BodyLimit:     1024 * 1024 * 50,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
	}))
	app.Use(logger.New())
	app.Use(recover.New())

	api := app.Group("/api")
	v1 := api.Group("/v1")

	settings := v1.Group("/settings")
	settings.Get("/", controller.GetSettings)
	settings.Post("/", controller.UpdateSettings)

	accounts := v1.Group("/account")
	accounts.Get("/", controller.GetAccounts)
	accounts.Post("/", controller.CreateAccount)
	accounts.Post("/delete", controller.DeleteAccounts)
	accounts.Post("/:id", controller.UpdateAccount)

	proxies := v1.Group("/proxy")
	proxyGroups := proxies.Group("/group")
	proxyGroups.Get("/", controller.GetProxyGroups)
	proxyGroups.Post("/", controller.CreateProxyGroup)
	proxyGroups.Post("/delete", controller.DeleteProxyGroups)
	proxyGroups.Post("/clone", controller.CloneProxyGroups)
	proxyGroups.Post("/:id", controller.UpdateProxyGroup)

	profiles := v1.Group("/profile")
	profileGroups := profiles.Group("/group")
	profileGroups.Get("/", controller.GetProfileGroups)
	profileGroups.Post("/", controller.CreateProfileGroup)
	profileGroups.Post("/delete", controller.DeleteProfileGroups)
	profileGroups.Post("/clone", controller.CloneProfileGroups)
	profileGroups.Post("/addProfiles", controller.AddProfilesToGroups)
	profileGroups.Post("/removeProfiles", controller.RemoveProfilesFromGroups)
	profileGroups.Post("/:id", controller.UpdateProfileGroup)

	profiles.Get("/", controller.GetProfiles)
	profiles.Post("/", controller.CreateProfile)
	profiles.Post("/delete", controller.DeleteProfiles)
	profiles.Post("/clone", controller.CloneProfiles)
	profiles.Post("/import", controller.ImportProfiles)
	profiles.Post("/export", controller.ExportProfiles)
	profiles.Post("/:id", controller.UpdateProfile)

	tasks := v1.Group("/task")
	tasks.Post("/", controller.CreateTasks)
	// tasks.Post("/update", controller.UpdateTasks)
	tasks.Post("/delete", controller.DeleteTasks)
	// 		TEST AFTER START/STOP FUNCTIONALITY
	// tasks.Post("/clone", controller.CloneTasks)
	// tasks.Post("/start", controller.StartTasks)
	// tasks.Post("/stop", controller.StopTasks)

	taskGroups := tasks.Group("/group")
	taskGroups.Get("/", controller.GetTaskGroups)
	taskGroups.Post("/", controller.CreateTaskGroup)
	// taskGroups.Post("/update", controller.UpdateTaskGroups)
	taskGroups.Post("/delete", controller.DeleteTaskGroups)
	// 		TEST AFTER START/STOP FUNCTIONALITY
	// taskGroups.Post("/clone", controller.CloneTaskGroups)
	// taskGroups.Post("/start", controller.StartTaskGroups)
	// taskGroups.Post("/stop", controller.StopTaskGroups)

	checkouts := v1.Group("/checkouts")
	checkouts.Get("/", controller.GetCheckouts)

	misc := v1.Group("/misc")
	misc.Post("/version", controller.SetVersion)
	misc.Post("/discord/test", controller.TestDiscord)

	go app.Listen(":10000")
}

func GetApp() *fiber.App {
	return app
}
