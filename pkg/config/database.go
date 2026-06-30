package config

import (
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// Global clients holding the active connections
var (
	MongoClient *mongo.Database
	GormClient  *gorm.DB
)

type DBType string

const (
	DBTypeMongo DBType = "mongo"
	DBTypeGorm  DBType = "gorm"
)

// DBConnection is a strictly typed wrapper for database connections.
type DBConnection interface {
	Type() DBType
	Mongo() *mongo.Database
	Gorm() *gorm.DB
}

type mongoConnection struct {
	db *mongo.Database
}

func NewMongoConnection(db *mongo.Database) DBConnection {
	return &mongoConnection{db: db}
}

func (m *mongoConnection) Type() DBType {
	return DBTypeMongo
}

func (m *mongoConnection) Mongo() *mongo.Database {
	return m.db
}

func (m *mongoConnection) Gorm() *gorm.DB {
	return nil
}

type gormConnection struct {
	db *gorm.DB
}

func NewGormConnection(db *gorm.DB) DBConnection {
	return &gormConnection{db: db}
}

func (g *gormConnection) Type() DBType {
	return DBTypeGorm
}

func (g *gormConnection) Mongo() *mongo.Database {
	return nil
}

func (g *gormConnection) Gorm() *gorm.DB {
	return g.db
}

// GetActiveConnection retrieves the active database connection
// based on which global client has been initialized.
func GetActiveConnection() DBConnection {
	if MongoClient != nil {
		return NewMongoConnection(MongoClient)
	}
	if GormClient != nil {
		return NewGormConnection(GormClient)
	}
	return nil
}
