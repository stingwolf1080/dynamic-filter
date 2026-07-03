package mongox

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
)

// Migrate creates all the parsed indexes for the registered models in the provided MongoDB database.
func Migrate(ctx context.Context, db *mongo.Database) error {
	models := GetAllModels()
	
	for _, m := range models {
		if len(m.Index) == 0 {
			continue
		}
		
		collection := db.Collection(m.Table)
		
		log.Printf("[migration] creating %d indexes for collection: %s", len(m.Index), m.Table)
		
		_, err := collection.Indexes().CreateMany(ctx, m.Index)
		if err != nil {
			return fmt.Errorf("failed to create indexes for table %s: %w", m.Table, err)
		}
		
		log.Printf("[migration] successfully created indexes for collection: %s", m.Table)
	}
	
	return nil
}
