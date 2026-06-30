package config

import (
	"context"
	"database/sql"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

// Global clients holding the active connections
var (
	MongoClient *mongo.Client
	GormClient  *gorm.DB
	SqlcClient  *sql.DB
)

type DBType string

const (
	DBTypeMongo DBType = "mongo"
	DBTypeGorm  DBType = "gorm"
	DBTypeSqlc  DBType = "sqlc"
)

// DBConnection is a strictly typed wrapper for database connections.
type DBConnection interface {
	Type() DBType
	Mongo() *mongo.Client
	Gorm() *gorm.DB
	Sqlc() *sql.DB
}

type mongoConnection struct {
	db *mongo.Client
}

func NewMongoConnection(db *mongo.Client) DBConnection {
	return &mongoConnection{db: db}
}

func (m *mongoConnection) Type() DBType {
	return DBTypeMongo
}

func (m *mongoConnection) Mongo() *mongo.Client {
	return m.db
}

func (m *mongoConnection) Gorm() *gorm.DB {
	return nil
}

func (m *mongoConnection) Sqlc() *sql.DB {
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

func (g *gormConnection) Mongo() *mongo.Client {
	return nil
}

func (g *gormConnection) Gorm() *gorm.DB {
	return g.db
}

func (g *gormConnection) Sqlc() *sql.DB {
	return nil
}

type sqlcConnection struct {
	db *sql.DB
}

func NewSqlcConnection(db *sql.DB) DBConnection {
	return &sqlcConnection{db: db}
}

func (s *sqlcConnection) Type() DBType {
	return DBTypeSqlc
}

func (s *sqlcConnection) Mongo() *mongo.Client {
	return nil
}

func (s *sqlcConnection) Gorm() *gorm.DB {
	return nil
}

func (s *sqlcConnection) Sqlc() *sql.DB {
	return s.db
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
	if SqlcClient != nil {
		return NewSqlcConnection(SqlcClient)
	}
	return nil
}

// ConnectMongo initializes the global MongoClient using a connection URI.
func ConnectMongo(ctx context.Context, uri string) error {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}
	
	// Verify the connection
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}
	
	MongoClient = client
	return nil
}

// ConnectGorm initializes the global GormClient using a specific dialector (e.g., postgres.Open(dsn)).
func ConnectGorm(dialector gorm.Dialector, cfg *gorm.Config) error {
	if cfg == nil {
		cfg = &gorm.Config{}
	}
	db, err := gorm.Open(dialector, cfg)
	if err != nil {
		return err
	}
	
	GormClient = db
	return nil
}

// ConnectSqlc initializes the global SqlcClient using standard database/sql driver.
func ConnectSqlc(driverName, dataSourceName string) error {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return err
	}
	
	// Verify connection
	if err := db.Ping(); err != nil {
		return err
	}
	
	SqlcClient = db
	return nil
}
