package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"os"
	"reflect"
	"restaurant-management-backend/internal/types"
	"strconv"
	"strings"
	"time"
)

// Service represents a service that interacts with a database.
type Service interface {
	Health() map[string]string
	GetDB() *sqlx.DB
	GetUserByID(id string) (*types.User, error)
	CreateUser(user types.User) (*types.User, error)
	UpdateUser(user types.User, id string) (*types.User, error)
	DeleteUser(id string) error
	ListUsers() ([]types.User, error)
	Close() error
}

type service struct {
	db *sqlx.DB
}

var (
	database   = os.Getenv("DB_DATABASE")
	password   = os.Getenv("DB_PASSWORD")
	username   = os.Getenv("DB_USERNAME")
	port       = os.Getenv("DB_PORT")
	host       = os.Getenv("DB_HOST")
	schema     = os.Getenv("DB_SCHEMA")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
	db, err := sqlx.Connect("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = &service{
		db: db,
	}
	mig, err := migrate.New(
		"file:///Users/krym/GolandProjects/restaurant-management-sadeem/restaurant-management-backend/internal/database/migrations",
		os.Getenv("DATABASE_URL"),
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := mig.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}
		log.Printf("migrations: %v", err.Error())
	}
	return dbInstance
}

func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}

func (s *service) GetDB() *sqlx.DB {
	return s.db
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

func UpdateBUILDER(data interface{}, id string, suffix ...string) (string, []interface{}, error) {
	v := reflect.ValueOf(data)
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	if t.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("data must be a struct")
	}

	// add s to type name user = users, vendor = vendors
	tableName := strings.ToLower(t.Name()) + "s"

	columns := []string{}
	values := []interface{}{}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag != "" {
			if !v.Field(i).IsZero() {
				columns = append(columns, dbTag)
				values = append(values, v.Field(i).Interface())
			}
		}
	}

	if len(columns) == 0 {
		return "", nil, fmt.Errorf("struct is empty")
	}

	updateBuilder := QB.Update(tableName).Where(squirrel.Eq{"id": id})

	for i := 0; i < len(columns); i++ {
		updateBuilder = updateBuilder.Set(columns[i], values[i])
	}

	if len(suffix) > 0 {
		updateBuilder = updateBuilder.Suffix(suffix[0])
	}

	return updateBuilder.ToSql()

}

func InsertBUILDER(data interface{}, suffix ...string) (string, []interface{}, error) {
	v := reflect.ValueOf(data)
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	if t.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("data must be a struct")
	}

	// add s to type name user = users, vendor = vendors
	tableName := strings.ToLower(t.Name()) + "s"

	columns := []string{}
	values := []interface{}{}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag != "" {
			if !v.Field(i).IsZero() {
				columns = append(columns, dbTag)
				values = append(values, v.Field(i).Interface())
			}
		}
	}

	if len(columns) == 0 {
		return "", nil, fmt.Errorf("struct is empty")
	}

	var insertBuilder squirrel.InsertBuilder
	if len(suffix) > 0 {
		insertBuilder = QB.Insert(tableName).
			Columns(columns...).
			Values(values...).
			Suffix(suffix[0])
	} else {
		insertBuilder = QB.Insert(tableName).
			Columns(columns...).
			Values(values...)
	}
	return insertBuilder.ToSql()
}

func deleteById(s *service, id string, table string, suffix ...string) (*string, error) {

	var data *string
	var deleteBuilder squirrel.DeleteBuilder
	if len(suffix) > 0 {
		deleteBuilder = QB.Delete(table).Where(squirrel.Eq{"id": id}).Suffix(suffix[0])
	} else {
		deleteBuilder = QB.Delete(table).Where(squirrel.Eq{"id": id})
	}
	query, args, err := deleteBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error deleting user failed building sql query: %w", err)
	}
	if len(suffix) > 0 {
		err = s.db.QueryRowx(query, args...).Scan(&data)
		if err != nil {
			return nil, fmt.Errorf("error deleting user failed query: %w", err)
		}
		return data, nil

	} else {
		result, err := s.db.Exec(query, args...)
		if err != nil {
			return nil, fmt.Errorf("error deleting user failed sql exec: %w", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return nil, fmt.Errorf("error deleting user failed to fetch rows affected:%w", err)
		}
		if affected == 0 {
			return nil, fmt.Errorf("error deleting user: no rows affected")
		}

	}
	return nil, nil
}
