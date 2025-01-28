package main

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"main/constants"
	"main/controllers"
	"main/models"
	"net/http"
	"os"
)

func startDatabase() *gorm.DB {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		return nil
	}

	host := os.Getenv("HOST")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	port := os.Getenv("DATABASE_PORT")

	envVariables := []string{host, user, password, dbname, port}

	for _, envVar := range envVariables {
		if envVar == constants.EMPTY {
			log.Fatal("One or more database environment variables are not set")
		}
	}

	// Database connection string
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, dbname, port)
	// Open the database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}

	// Enable uuid-ossp extension
	err = db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error
	if err != nil {
		log.Fatalf("failed to enable uuid-ossp extension: %v", err)
	}

	migrateSchemas(db)

	return db
}

func migrateSchemas(db *gorm.DB) {
	err := db.AutoMigrate(&models.Post{}, &models.Follow{}, &models.Like{}, &models.User{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
}

func startServer() {
	db := startDatabase()

	if db == nil {
		fmt.Println("Error starting the database")
		return
	}

	s, serverError := db.DB()
	if serverError != nil {
		return
	}

	// Defer its closing
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			return
		}
	}(s)

	// Here should go the functions for each endpoint
	// TODO: Implement a function that activates all the endpoints

	http.HandleFunc(controllers.UserSignUpEndpoint.Path, func(writer http.ResponseWriter, request *http.Request) {
		controllers.UserSignUpEndpoint.HandlerFunction(writer, request, db)
	})

	http.HandleFunc(controllers.UserLoginEndpoint.Path, func(writer http.ResponseWriter, request *http.Request) {
		controllers.UserLoginEndpoint.HandlerFunction(writer, request, db)
	})

	http.HandleFunc(controllers.SearchUserEndpoint.Path, func(writer http.ResponseWriter, request *http.Request) {
		controllers.SearchUserEndpoint.HandlerFunction(writer, request, db)
	})

	http.HandleFunc(controllers.SearchPostEndpoint.Path, func(writer http.ResponseWriter, request *http.Request) {
		controllers.SearchPostEndpoint.HandlerFunction(writer, request, db)
	})

	http.HandleFunc(controllers.CreatePostEndpoint.Path, func(writer http.ResponseWriter, request *http.Request) {
		controllers.CreatePostEndpoint.HandlerFunction(writer, request, db)
	})

	http.HandleFunc(controllers.GetSpecificPostEndpoint.Path, func(writer http.ResponseWriter, request *http.Request) {
		controllers.GetSpecificPostEndpoint.HandlerFunction(writer, request, db)
	})

	http.HandleFunc(controllers.EditPostEndpoint.Path, func(writer http.ResponseWriter, request *http.Request) {
		controllers.EditPostEndpoint.HandlerFunction(writer, request, db)
	})

	http.HandleFunc(controllers.DeletePostEndpoint.Path, func(writer http.ResponseWriter, request *http.Request) {
		controllers.DeletePostEndpoint.HandlerFunction(writer, request, db)
	})

	http.HandleFunc(controllers.GetAllPostsEndpoint.Path, func(writer http.ResponseWriter, request *http.Request) {
		controllers.GetAllPostsEndpoint.HandlerFunction(writer, request, db)
	})

	http.HandleFunc(controllers.GetAllPostsByUserIDEndpoint.Path, func(writer http.ResponseWriter, request *http.Request) {
		controllers.GetAllPostsByUserIDEndpoint.HandlerFunction(writer, request, db)
	})

	http.HandleFunc(controllers.ViewUserProfileEndpoint.Path, func(writer http.ResponseWriter, request *http.Request) {
		controllers.ViewUserProfileEndpoint.HandlerFunction(writer, request, db)
	})

	http.HandleFunc(controllers.ViewFollowersProfileEndpoint.Path,
		func(writer http.ResponseWriter, request *http.Request) {
			controllers.ViewFollowersProfileEndpoint.HandlerFunction(writer, request, db)
		})

	http.HandleFunc(controllers.ViewFollowingProfileEndpoint.Path,
		func(writer http.ResponseWriter, request *http.Request) {
			controllers.ViewFollowingProfileEndpoint.HandlerFunction(writer, request, db)
		})

	http.HandleFunc(controllers.EditUserProfileEndpoint.Path, func(writer http.ResponseWriter, request *http.Request) {
		controllers.EditUserProfileEndpoint.HandlerFunction(writer, request, db)
	})

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == constants.EMPTY {
		log.Panic("serverPort environment variable is not set")
	}

	fmt.Printf("Server running on port %s", serverPort)
	serverError = http.ListenAndServe(":"+serverPort, nil)
	if serverError != nil {
		return
	}
}

func main() {
	startServer()
}
