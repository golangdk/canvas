package integrationtest

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/maragudk/env"
	"github.com/maragudk/migrate"

	"canvas/storage"
)

var once sync.Once

// CreateDatabase for testing.
// Usage:
// 	db, cleanup := CreateDatabase()
// 	defer cleanup()
// 	…
func CreateDatabase() (*storage.Database, func()) {
	env.MustLoad("../.env-test")

	once.Do(initDatabase)

	db, cleanup := connect("postgres")
	defer cleanup()

	dropConnections(db.DB, "template1")

	name := env.GetStringOrDefault("DB_NAME", "test")
	dropConnections(db.DB, name)
	db.DB.MustExec(`drop database if exists ` + name)
	db.DB.MustExec(`create database ` + name)

	return connect(name)
}

func initDatabase() {
	db, cleanup := connect("template1")
	defer cleanup()

	for err := db.Ping(context.Background()); err != nil; {
		time.Sleep(100 * time.Millisecond)
	}

	if err := migrate.Up(context.Background(), db.DB.DB, os.DirFS("../storage/migrations")); err != nil {
		panic(err)
	}
	if err := migrate.Down(context.Background(), db.DB.DB, os.DirFS("../storage/migrations")); err != nil {
		panic(err)
	}
	if err := migrate.Up(context.Background(), db.DB.DB, os.DirFS("../storage/migrations")); err != nil {
		panic(err)
	}

	if err := db.DB.Close(); err != nil {
		panic(err)
	}
}

func connect(name string) (*storage.Database, func()) {
	db := storage.NewDatabase(storage.NewDatabaseOptions{
		Host:               env.GetStringOrDefault("DB_HOST", "localhost"),
		Port:               env.GetIntOrDefault("DB_PORT", 5432),
		User:               env.GetStringOrDefault("DB_USER", "test"),
		Password:           env.GetStringOrDefault("DB_PASSWORD", ""),
		Name:               name,
		MaxOpenConnections: 10,
		MaxIdleConnections: 10,
	})
	if err := db.Connect(); err != nil {
		panic(err)
	}
	return db, func() {
		if err := db.DB.Close(); err != nil {
			panic(err)
		}
	}
}

func dropConnections(db *sqlx.DB, name string) {
	db.MustExec(`
		select pg_terminate_backend(pg_stat_activity.pid)
		from pg_stat_activity
		where pg_stat_activity.datname = $1 and pid <> pg_backend_pid()`, name)
}
