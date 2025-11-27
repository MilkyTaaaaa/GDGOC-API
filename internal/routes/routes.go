package routes

import(
	"GDGOC-API/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// setup
func SetupRoutes(app *fiber.App, menuHandler *handlers.MenuHandler){
	app.Get("/health", func(c *fiber.Ctx) error{
		return c.JSON(fiber.Map{
			"status": "ok",
			"message": "menu API berjalan",
		})
	})

	setupMenuRoutes(app, menuHandler)
}

func setupMenuRoutes(router fiber.Router, handler *handlers.MenuHandler){
	// menu route
	router.Post("/menu/recommendations", handler.GetRecommendations)
	router.Get("/menu/group-by-category", handler.GroupByCategory)
	router.Get("/menu/search", handler.SearchMenus)
	router.Post("/menu", handler.CreateMenu)
	router.Get("/menu", handler.GetAllMenus)
	router.Get("/menu/:id", handler.GetMenuByID)
	router.Put("/menu/:id", handler.UpdateMenu)
	router.Delete("/menu/:id", handler.DeleteMenu)
	
}