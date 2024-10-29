package adapters

import (
	"gorm.io/driver/mysql"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "strings"
)

type SQLAdapter struct {
    db *gorm.DB
}

// NewSQLAdapter initializes a new SQLAdapter with a given DSN and optionally a database type.
// dbType can be "postgres" or "mysql", with "postgres" as the default.
func NewSQLAdapter(dsn string, dbType string) *SQLAdapter {
    var db *gorm.DB
    var err error

    // Choose the driver based on dbType
    switch strings.ToLower(dbType) {
    case "mysql":
        db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
    case "postgres":
        fallthrough // Use postgres as the default
    default:
        db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    }

    if err != nil {
        panic("Failed to connect to SQL database: " + err.Error())
    }

    return &SQLAdapter{db: db}
}
// Create inserts a new record into the database.
func (g *SQLAdapter) GetDB() *gorm.DB {
    return g.db
}

// Create inserts a new record into the database.
func (g *SQLAdapter) Create(model interface{}) error {
    return g.db.Create(model).Error
}

// Read retrieves a record by ID from the database.
func (g *SQLAdapter) Read(id uint, model interface{}) error {
    return g.db.First(model, "id = ?", id).Error
}

// Update modifies an existing record in the database.
func (g *SQLAdapter) Update(model interface{}) error {
    return g.db.Save(model).Error
}

// Delete removes a record by ID from the database.
func (g *SQLAdapter) Delete(id uint, model interface{}) error {
    return g.db.Delete(model, "id = ?", id).Error
}

// BeginTransaction starts a new transaction and returns a new SQLAdapter instance with the transactional DB.
func (g *SQLAdapter) BeginTransaction() (*SQLAdapter, error) {
    tx := g.db.Begin()
    if tx.Error != nil {
        return nil, tx.Error
    }
    return &SQLAdapter{db: tx}, nil
}

// Commit commits the transaction.
func (g *SQLAdapter) Commit() error {
    return g.db.Commit().Error
}

// Rollback rolls back the transaction.
func (g *SQLAdapter) Rollback() error {
    return g.db.Rollback().Error
}

// RawQuery executes a raw SQL query and scans the result into the provided destination.
func (g *SQLAdapter) RawQuery(query string, params []interface{}, dest interface{}) error {
    return g.db.Raw(query, params...).Scan(dest).Error
}

// Close terminates the database connection.
func (g *SQLAdapter) Close() error {
    db, err := g.db.DB()
    if err != nil {
        return err
    }
    return db.Close()
}
