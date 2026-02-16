package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

var DB *sql.DB
var RDB *redis.Client

func InitDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"), 
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("❌ Failed to open DB:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("❌ DB not reachable:", err)
	} // <--- health connection
	
	// Redis Ping <-> Pong

	log.Println("✅ Connected to PostgreSQL")
}
func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})
	if _, err := RDB.Ping(context.Background()).Result(); err != nil {
		log.Fatal("❌ Failed to connect to Redis:", err)
	}	
	log.Println("✅ Connected to Redis")
}