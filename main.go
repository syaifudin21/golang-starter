package main

import (
	"exam/database"
	"exam/internal/handler"
	"exam/internal/i18n"
	"exam/internal/middleware"
	"exam/internal/repository"
	"exam/internal/routes"
	"exam/internal/service"
	"exam/internal/websocket"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/casbin/casbin/v2"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <command>")
		fmt.Println("Commands: api, migrate")
		return
	}

	command := os.Args[1]

	switch command {
	case "api":
		runAPI()
	case "migrate":
		runMigration()
	default:
		fmt.Println("Unknown command:", command)
	}
}

func runAPI() {
	i18n.Init()
	db, err := database.NewDB()
	if err != nil {
		panic(err)
	}

	fmt.Println("Database migration and model sync complete")

	// Start the websocket hub
	hub := websocket.NewHub()
	go hub.Run()

	e := echo.New()

	// Enable CORS with a more explicit configuration
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.POST, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderXRequestedWith, echo.HeaderContentLength, echo.HeaderContentDisposition, echo.HeaderCacheControl, echo.HeaderSetCookie},
		AllowCredentials: true,
	}))

	// Serve static files from the 'uploads' directory
	e.Static("/uploads", "uploads")

	modelPath := filepath.Join("internal", "config", "rbac_model.conf")
	policyPath := filepath.Join("internal", "config", "policy.csv")
	enforcer, err := casbin.NewEnforcer(modelPath, policyPath)
	if err != nil {
		log.Fatalf("Failed to create Casbin enforcer: %v", err)
	}

	// Initialize Google OAuth2 config
	googleOauthConfig := &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	quizRepo := repository.NewQuizRepository(db)
	uploadedFileRepo := repository.NewUploadedFileRepository(db)

	// Initialize services
	deviceService := service.NewDeviceService(deviceRepo)
	authService := service.NewAuthService(userRepo, deviceRepo, googleOauthConfig.ClientID)
	quizService := service.NewQuizService(quizRepo, hub)
	fileService := service.NewFileService(uploadedFileRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, googleOauthConfig)
	accountHandler := handler.NewAccountHandler(authService, deviceService)
	userHandler := handler.NewUserHandler(authService)
	quizHandler := handler.NewQuizHandler(quizService)
	websocketHandler := handler.NewWebsocketHandler(hub, quizService)
	fileHandler := handler.NewFileHandler(fileService)

	// Register health check
	e.GET("/health", func(c echo.Context) error {
		sqlDB, err := db.DB()
		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to get db instance")
		}
		if err := sqlDB.Ping(); err != nil {
			return c.String(http.StatusInternalServerError, "database not connected")
		}
		return c.String(http.StatusOK, "database connected")
	})

	// Register routes
	routes.AuthRoutes(e, authHandler)

	v1 := e.Group("/api/v1")
	v1.Use(middleware.JWTAuthMiddleware(deviceRepo))
	v1.Use(middleware.CasbinAuthMiddleware(enforcer))
	routes.APIRoutes(v1, authHandler, accountHandler, userHandler, quizHandler, websocketHandler, fileHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}

func runMigration() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run . migrate <up|down|force|version>")
		return
	}

	db, err := database.NewDB()
	if err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB from GORM connection")
	}

	driver, err := mysql.WithInstance(sqlDB, &mysql.Config{})
	if err != nil {
		log.Fatalf("Could not start sql migration: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://database/migration",
		"mysql",
		driver,
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	subcommand := os.Args[2]
	switch subcommand {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("An error occurred while migrating up: %v", err)
		}
		fmt.Println("Migration up success")
	case "down":
		if len(os.Args) > 3 && os.Args[3] == "--all" {
			if err := m.Down(); err != nil && err != migrate.ErrNoChange {
				log.Fatalf("An error occurred while migrating down all: %v", err)
			}
			fmt.Println("Migration down all success")
		} else {
			if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
				log.Fatalf("An error occurred while migrating one step down: %v", err)
			}
			fmt.Println("Migration one step down success")
		}
	case "force":
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run . migrate force <version>")
			return
		}
		version, err := strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatal("Invalid version")
		}
		if err := m.Force(version); err != nil {
			log.Fatalf("An error occurred while forcing migration: %v", err)
		}
		fmt.Println("Migration force success")
	default:
		fmt.Println("Unknown subcommand:", subcommand)
	}
}
