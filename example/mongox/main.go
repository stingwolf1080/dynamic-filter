package main

import (
	"fmt"
	"log"

	"github.com/stingwolf1080/dynamic-filter/pkg/db"
	"github.com/stingwolf1080/dynamic-filter/pkg/db/mongox"
	"github.com/stingwolf1080/dynamic-filter/repository"
)

type User struct {
	ID    string `bson:"_id,omitempty" json:"id"`
	Name  string `bson:"name" json:"name"`
	Email string `bson:"email" json:"email"`
	Age   int    `bson:"age" json:"age"`
}

func main() {
	// 1. Initialize MongoDB Connection
	fmt.Println("Connecting to MongoDB...")
	client := &mongox.Conn{}
	err := client.Connect(db.Options{
		URI: "mongodb://localhost:27017",
		Database: "test_db",
	})
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Close()
	fmt.Println("Connected successfully!")

	// 2. Initialize the Generic Repository for the User model
	// We pass the connected client as the db.Connection
	userRepo := repository.NewRepository[User](client, nil, nil)

	// 3. Create a new User
	newUser := User{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
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

	// Because GetByFilter returns a types.Message where Data is the slice of Users
	users, ok := readMsg.Data.(*[]User)
	if !ok {
		log.Printf("Expected *[]User, got %T", readMsg.Data)
	} else {
		for i, u := range *users {
			fmt.Printf("Result %d: %+v\n", i, u)
		}
	}
}
