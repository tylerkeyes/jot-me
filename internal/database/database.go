package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// WriteNote inserts the given 'note' into the 'groupName' table.
	// If the 'groupName' table does not exist, it will be created.
	// Any errors will be returned, else nil.
	WriteNote(groupName string, note string) error

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error
}

type service struct {
	db *sql.DB
}

var (
	dburl            = os.Getenv("DB_URL")
	dbInstance       *service
	defaultTableName = "general"
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	db, err := sql.Open("sqlite3", dburl)
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal(err)
	}

	dbInstance = &service{
		db: db,
	}

	// create the 'general' table for uncategorized notes,
	// and the '_group_names' table for tracking enabled categories
	fmt.Println("initalizing db")
	dbInstance.CreateNamesTable()
	dbInstance.CreateGroupTable(defaultTableName)
	return dbInstance
}

// WriteNote inserts the given 'note' into the 'groupName' table.
// If the 'groupName' table does not exist, it will be created.
// Any errors will be returned, else nil.
func (s *service) WriteNote(groupName string, note string) error {
	table := groupName
	if table == "" {
		table = defaultTableName
	}

	// create the new table if it doesn't exist
	if !s.CheckTableExists(table) {
		fmt.Printf("creating new table: %v\n", table)
		s.CreateGroupTable(table)
	}

	queryStr := `INSERT INTO %s (note) VALUES ('%s');`
	query := fmt.Sprintf(queryStr, table, note)

	_, err := s.db.Exec(query)
	if err != nil {
		fmt.Printf("could not save the note in the group %v\n", groupName)
		return err
	}

	return nil
}

// CheckTableExists checks if the tableName exists as a table.
func (s *service) CheckTableExists(tableName string) bool {
	query := "SELECT name FROM sqlite_master WHERE type='table' and name=?"
	var name string
	err := s.db.QueryRow(query, tableName).Scan(&name)
	return err == nil
}

// CreateGroupTable creates a table with the name groupName.
func (s *service) CreateGroupTable(groupName string) {
	// create the default note group
	query := `
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		note TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	createDefaultGroup := fmt.Sprintf(query, groupName)
	_, err := s.db.Exec(createDefaultGroup)
	if err != nil {
		fmt.Printf("could not create the %v note group. %v\n", groupName, err)
		os.Exit(1)
	}

	s.AddGroupName(groupName)
}

// AddGroupName adds the new group to the stored list of groups
func (s *service) AddGroupName(groupName string) {
	groupNameExists := `SELECT group_name FROM _group_names WHERE group_name=?`
	var expected string
	err := s.db.QueryRow(groupNameExists, groupName).Scan(&expected)
	if err == nil {
		return
	}

	query := `INSERT INTO _group_names (group_name) VALUES (?);`
	_, err = s.db.Exec(query, groupName)
	if err != nil {
		fmt.Printf("could not save the new group\n")
	}
}

// CreateNamesTable creates the '_group_names' table.
// The '_group_names' table is used to track all groups used for categorizing notes.
func (s *service) CreateNamesTable() {
	query := `
	CREATE TABLE IF NOT EXISTS _group_names (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_name TEXT NOT NULL
	);`

	_, err := s.db.Exec(query)
	if err != nil {
		fmt.Printf("could not create the _group_names table: %v\n", err)
		os.Exit(1)
	}
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf(fmt.Sprintf("db down: %v", err)) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	return s.db.Close()
}
