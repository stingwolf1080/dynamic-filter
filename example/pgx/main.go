package main

import (
	"fmt"
	"log"

	"github.com/stingwolf1080/dynamic-filter/pkg/db"
	"github.com/stingwolf1080/dynamic-filter/pkg/db/pgx"
	"github.com/stingwolf1080/dynamic-filter/repository"
)

type User struct {
	ID    string `db:"id" json:"id"`
	Name  string `db:"name" json:"name"`
	Email string `db:"email" json:"email"`
	Age   int    `db:"age" json:"age"`
}

func main() {
	// 1. Initialize Postgres Connection via pgx
	fmt.Println("Connecting to PostgreSQL...")
	client := &pgx.Conn{}
	err := client.Connect(db.Options{
		// Replace with your actual Postgres URI
		URI: "postgres://postgres:password@localhost:5432/test_db?sslmode=disable",
	})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer client.Close()
	fmt.Println("Connected successfully!")

	// 2. Initialize the Generic Repository for the User model
	// This will resolve the table name automatically (e.g. "users")
	userRepo := repository.NewRepository[User](client, nil, nil)

	// 3. Create a new User
	newUser := User{
		ID:    "123-abc", // Normally you'd generate a UUID or let Postgres handle it
		Name:  "Jane Doe",
		Email: "jane@example.com",
		Age:   28,
	}

	fmt.Println("Creating user...")
	msg := userRepo.Create(newUser)
	if msg.Status == "error" {
		log.Fatalf("Error creating user: %v", msg.Message)
	}
	fmt.Println("User created successfully!")

	// 4. Read User back using dynamic filters
	// Here we query for Age >= 25
	fmt.Println("Querying for users with age >= 25...")
	filterStr := `filter[age]=[gte]:25`
	readMsg := userRepo.GetByFilter(filterStr)
	if readMsg.Status == "error" {
		log.Fatalf("Error querying users: %v", readMsg.Message)
	}

	users, ok := readMsg.Data.(*[]User)
	if !ok {
		log.Printf("Expected *[]User, got %T", readMsg.Data)
	} else {
		for i, u := range *users {
			fmt.Printf("Result %d: %+v\n", i, u)
		}
	}
}
