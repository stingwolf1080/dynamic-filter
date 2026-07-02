package main

import (
	"fmt"
	"log"
	"os"

	"github.com/stingwolf1080/dynamic-filter/pkg/db"
	"github.com/stingwolf1080/dynamic-filter/pkg/db/sqlx"
	"github.com/stingwolf1080/dynamic-filter/pkg/helper"
	"github.com/stingwolf1080/dynamic-filter/pkg/util/types"
	"github.com/stingwolf1080/dynamic-filter/repository"
)

type User struct {
	ID    string `db:"id" json:"id"`
	Name  string `db:"name" json:"name"`
	Email string `db:"email" json:"email"`
	Age   int    `db:"age" json:"age"`
}

func main() {
	dbFile := "test.db"
	os.Remove(dbFile) // ensure clean start

	// 1. Initialize SQLite Connection via sqlx
	fmt.Println("Connecting to SQLite...")
	client := &sqlx.Conn{}
	err := client.Connect(db.Options{
		URI: dbFile,
	})
	if err != nil {
		log.Fatalf("Failed to connect to SQLite: %v", err)
	}
	defer func() {
		client.Close()
		os.Remove(dbFile) // cleanup
	}()
	fmt.Println("Connected successfully!")

	tableName := helper.GetModelTableGeneric[User]()
	fmt.Println("Table Name: ", tableName)

	_, err = client.DB.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			name TEXT,
			email TEXT,
			age INTEGER
		)
	`, tableName))
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// 2. Initialize the Generic Repository for the User model
	userRepo := repository.NewRepository[User](client, nil, nil)

	// 3. Create a few Users
	usersToCreate := []User{
		{ID: "1", Name: "Alice Bob", Email: "alice@example.com", Age: 22},
		{ID: "2", Name: "Charlie", Email: "charlie@example.com", Age: 28},
		{ID: "3", Name: "Dave", Email: "dave@example.com", Age: 35},
		{ID: "4", Name: "Eve", Email: "eve@example.com", Age: 24},
	}

	fmt.Println("Creating users...")
	for _, u := range usersToCreate {
		msg := userRepo.Create(u)
		if msg.Status == "error" {
			log.Fatalf("Error creating user %s: %v, %v", u.Name, msg.Message, msg.MessageErr)
		}
	}
	fmt.Println("Users created successfully!\n")

	// 4. Read Users using complex dynamic filters
	fmt.Println("Querying for users with (age >= 24) AND (name == 'Alice Bob' OR name == 'Dave') sorted by -age...")
	filterStr := `filter[age]=[gte]:24&filter[name,name]=[eq]:Alice Bob,[eq]:Dave&sort[age]=desc`
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

	// 5. Update users using dynamic filter
	fmt.Println("\nUpdating users where age < 25 (Set Age to 25)...")
	updateMsg := userRepo.UpdateMany(`filter[age]=[lt]:25`, map[string]any{"age": 25})
	if updateMsg.Status == "error" {
		log.Fatalf("Error updating users: %v", updateMsg.Message)
	}
	fmt.Println("Update successful!")

	// 6. Delete users using dynamic filter
	fmt.Println("\nDeleting user with name 'Charlie'...")
	deleteMsg := userRepo.DeleteMany(`filter[name]=[eq]:Charlie`, types.DeletePost{})
	if deleteMsg.Status == "error" {
		log.Fatalf("Error deleting user: %v", deleteMsg.Message)
	}
	fmt.Println("Delete successful!")
}
