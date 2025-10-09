package main

import (
	"fmt"
	"net/http"

	"hariomtransport/config"
	"hariomtransport/db"
	"hariomtransport/db/mongo"
	"hariomtransport/db/postgres"
	"hariomtransport/handlers"
	"hariomtransport/repository"
	"hariomtransport/routes"
)

func main() {
	// Load config from .env or config file
	cfg := config.LoadConfig()

	// Run migrations (for Postgres)
	db.RunMigrations()

	var biltyRepo repository.BiltyRepository
	var userRepo repository.UserRepository
	var initialRepo repository.InitialRepository

	switch cfg.DBType {
	case "postgres":
		pg := postgres.NewPostgresDB(cfg.PostgresURL)
		if err := pg.Connect(); err != nil {
			panic(err)
		}
		defer pg.Disconnect()

		biltyRepo = repository.NewPostgresBiltyRepo(pg.Conn)
		userRepo = repository.NewPostgresUserRepo(pg.Conn)
		initialRepo = repository.NewPostgresInitialRepo(pg.Conn)

	case "mongo":
		mg := mongo.NewMongoDB(cfg.MongoURL)
		if err := mg.Connect(); err != nil {
			panic(err)
		}
		defer mg.Disconnect()

		biltyRepo = repository.NewMongoBiltyRepo(mg.Client)
		userRepo = repository.NewMongoUserRepo(mg.Client)
		initialRepo = repository.NewMongoInitialRepo(mg.Client)

	default:
		panic("DB_TYPE not supported")
	}

	// Handlers
	biltyHandler := &handlers.BiltyHandler{Repo: biltyRepo}
	userHandler := &handlers.UserHandler{Repo: userRepo}
	initialHandler := &handlers.InitialHandler{Repo: initialRepo}

	// PDF handler with combined repository
	pdfRepo := &repository.PDFRepository{
		BiltyRepo:   biltyRepo,
		InitialRepo: initialRepo,
	}
	pdfHandler := &handlers.PDFHandler{Repo: pdfRepo}

	// Setup routes including PDF
	routes.SetupRoutes(userHandler, biltyHandler, initialHandler, pdfHandler)

	port := cfg.Port
	fmt.Printf("Server running on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}
