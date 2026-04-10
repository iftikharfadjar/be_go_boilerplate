package main

import (
	"database/sql"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/proxy"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	"boilerplate/internal/db"
	"boilerplate/internal/domain"
	"boilerplate/internal/graphql"
	"boilerplate/internal/rest"
	"boilerplate/internal/repository/pocketbase"
	"boilerplate/internal/repository/sql_adapter"
	usecase_auth "boilerplate/internal/usecase/auth"
	"boilerplate/pkg/config"
)

func runMigrations(dbConn *sql.DB, dbType string, connString string) {
	log.Println("Running database migrations...")
	
	var driver database.Driver
	var err error

	if dbType == "postgres" {
		driver, err = postgres.WithInstance(dbConn, &postgres.Config{})
	} else if dbType == "sqlite" {
		driver, err = sqlite.WithInstance(dbConn, &sqlite.Config{})
	}

	if err != nil {
		log.Fatalf("could not instantiate migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://sql/migrations", dbType, driver)
	if err != nil {
		log.Fatalf("migration init failed: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("an error occurred while syncing the database.. %v", err)
	}
	log.Println("Database migrated successfully!")
}

func main() {
	// 1. Load Primary Config (config.json -> env vars)
	cfg := config.LoadConfig()
	log.Printf("Booting with DB_TYPE: %s", cfg.DBType)

	var authRepo domain.AuthRepository

	// 2. Conditional Dependency Injection
	switch cfg.DBType {
	case "postgres":
		dbConn, err := sql.Open("postgres", cfg.DBConnString)
		if err != nil {
			log.Fatal(err)
		}
		runMigrations(dbConn, "postgres", cfg.DBConnString)
		authRepo = sql_adapter.NewAuthRepository(dbConn, cfg.JWTSecret)

	case "sqlite":
		dbConn, err := sql.Open("sqlite", cfg.DBConnString)
		if err != nil {
			log.Fatal(err)
		}
		runMigrations(dbConn, "sqlite", cfg.DBConnString)
		authRepo = sql_adapter.NewAuthRepository(dbConn, cfg.JWTSecret)

	case "pocketbase":
		fallthrough
	default:
		pbApp := db.Init()
		authRepo = pocketbase.NewAuthRepository(pbApp)

		// Start PocketBase asynchronously
		go func() {
			if err := pbApp.Start(); err != nil {
				log.Fatalf("PocketBase start failed: %v", err)
			}
		}()
	}

	// 3. Initialize UseCases (Business Logic)
	authUseCase := usecase_auth.NewAuthUseCase(authRepo)

	// 4. Initialize Fiber app
	app := fiber.New()

	// 5. Setup REST Routes (Delivery)
	authHandler := rest.NewAuthHandler(authUseCase)
	authHandler.SetupRoutes(app)

	// 6. Setup GraphQL Routes (Delivery)
	graphql.SetupRoutes(app)

	// 7. Mount PocketBase Admin UI and Core API via Proxy (only applies if pb wrapper is reachable)
	if cfg.DBType == "pocketbase" {
		app.All("/_/*", proxy.Forward("http://127.0.0.1:8090/_/"))
		app.All("/api/*", proxy.Forward("http://localhost:8090/api/"))
	}

	// 8. Start Fiber Server
	log.Printf("Starting Fiber Server on :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("Fiber server failed: %v", err)
	}
}
