package main

import (
	"exam/database"
	"exam/internal/handler"
	"exam/internal/i18n"
	"exam/internal/middleware"
	"exam/internal/repository"
	"exam/internal/routes"
	"exam/internal/service"
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

	// Run migrations
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://database/migration",
		"mysql",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	fmt.Println("Migration check complete")

	e := echo.New()

	// Initialize Casbin enforcer
	modelPath := filepath.Join("internal", "config", "rbac_model.conf")
	policyPath := filepath.Join("internal", "config", "policy.csv")
	enforcer, err := casbin.NewEnforcer(modelPath, policyPath)
	if err != nil {
		log.Fatalf("Failed to create Casbin enforcer: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)

	// Initialize services and handlers
	deviceService := service.NewDeviceService(deviceRepo)
	authService := service.NewAuthService(userRepo, deviceRepo)
	authHandler := handler.NewAuthHandler(authService)
	accountHandler := handler.NewAccountHandler(db, authService, deviceService)
	userHandler := handler.NewUserHandler(authService)

	// Register routes
	e.GET("/health", func(c echo.Context) error {
		err := db.Ping()
		if err != nil {
			return c.String(http.StatusInternalServerError, "database not connected")
		}
		return c.String(http.StatusOK, "database connected")
	})

	routes.AuthRoutes(e, authHandler)

	v1 := e.Group("/api/v1")
	v1.Use(middleware.JWTAuthMiddleware(deviceRepo))
	v1.Use(middleware.CasbinAuthMiddleware(enforcer))
	routes.APIRoutes(v1, authHandler, accountHandler, userHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}

func runMigration() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run . migrate <subcommand>")
		fmt.Println("Subcommands: up, down, force")
		return
	}

	db, err := database.NewDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://database/migration",
		"mysql",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	subcommand := os.Args[2]
	switch subcommand {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		fmt.Println("Migration up success")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		fmt.Println("Migration down success")
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
			log.Fatal(err)
		}
		fmt.Println("Migration force success")
	default:
		fmt.Println("Unknown subcommand:", subcommand)
	}
}
