package main

import(
	"GDGOC-API/internal/config"
	"GDGOC-API/internal/database"
	"GDGOC-API/internal/gemini"
	"GDGOC-API/internal/handlers"
	"GDGOC-API/internal/repositories"
	"GDGOC-API/internal/routes"
	"GDGOC-API/internal/services"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main(){
	// loading
	log.Println("Loading konfigurasi...")
	config.LoadConfig()

	// koneksi db
	log.Println("Menghubungkan ke database...")
	database.ConnectDatabase()
	defer database.CloseDatabase()

	// ✅ DEKLARASI VARIABLE DI SCOPE YANG SAMA
	var geminiService *gemini.Service

    if config.GetConfig().GeminiAPIKey != "" {
        log.Println("Inisialisasi Gemini AI...")
        
        // ✅ GUNAKAN = BUKAN := (karena sudah dideklarasikan)
        var err error
        geminiService, err = gemini.NewService(config.GetConfig().GeminiAPIKey)
        if err != nil {
            log.Printf("Gagal inisialisasi Gemini: %v", err)
            geminiService = nil
        } else {
            log.Println("Gemini AI berhasil diinisialisasi")
        }
    } else {
        log.Println("Gemini API key belum di set - berjalan tanpa fitur AI")
        geminiService = nil
    }

	// layer app
	log.Println("Inisialisasi layer...")

	menuRepo := repositories.NewMenuRepository(database.GetDB())
	menuService := services.NewMenuService(menuRepo)
	
	// ✅ SEKARANG geminiService SUDAH TERDEFINISI DI SCOPE INI
	menuHandler := handlers.NewMenuHandler(menuService, geminiService) 

	log.Println("Creating Fiber app...")
	app := fiber.New(fiber.Config{
		AppName: "Menu Catalog API",
		ServerHeader: "Fiber",
		ErrorHandler: customErrorHandler,
	})

	// setup route
	log.Println("Setting route...")
	routes.SetupRoutes(app, menuHandler)

	// middleware
	setupMiddleware(app)

	
	// start
	port := config.GetConfig().Port
	log.Println("Server siap digunakan")
	log.Printf("Server dijalankan di http://localhost:%s\n", port)
	log.Printf("API Base : http://localhost:%s\n", port)
	log.Printf("Health check: http://localhost:%s/health\n", port)
	log.Printf("AI Recommendations: POST http://localhost:%s/menu/recommendations\n", port)

	// shutdown
	go func() {
		if err := app.Listen(":" + port); err != nil{
			log.Fatalf("Server gagal dijalankan: %v", err)
		}
	}()

	// interupsi
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Mematikan server...")
	if err := app.Shutdown(); err != nil{
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server berhasil dimatikan")
}

// setup middleware
func setupMiddleware(app *fiber.App){
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	// logger
	if config.GetConfig().AppEnv == "development"{
		app.Use(logger.New(logger.Config{
			Format:	"[${time}] ${status} - ${method} ${path} (${latency})\n",
			TimeFormat:	"15:04:05",
			TimeZone:	"Asia/Jakarta",
		}))
	}

	// CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// 404 handler
	app.Use(func(c *fiber.Ctx) error{
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Endpoint tidak ditemukan",
			"path": c.Path(),
		})
	})
}

// error handler
func customErrorHandler(c *fiber.Ctx, err error) error{
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok{
		code = e.Code
	}

	log.Printf("Error : %v", err)

	//return
	return c.Status(code).JSON(fiber.Map{
		"message": err.Error(),
		"status": code,
	})
}