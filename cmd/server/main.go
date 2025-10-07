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

	biltyHandler := &handlers.BiltyHandler{Repo: biltyRepo}
	userHandler := &handlers.UserHandler{Repo: userRepo}
	initialHandler := &handlers.InitialHandler{Repo: initialRepo}

	// Setup routes for users, bilty, and initial
	routes.SetupRoutes(userHandler, biltyHandler, initialHandler)

	port := cfg.Port
	fmt.Printf("Server running on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
