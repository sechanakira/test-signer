package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	var err error
	db, err = initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	createTableSQL := `
    CREATE TABLE IF NOT EXISTS user_session (
        id SERIAL PRIMARY KEY,
        user_id VARCHAR(255) NOT NULL,
        signature TEXT NOT NULL,
        answers JSONB NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create user_session table: %v", err)
	}
}

func initDB() (*sql.DB, error) {
	dsn := getDSN()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

func getDSN() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSL_MODE")

	return "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=" + sslmode
}

func FetchSessionData(userID, signature string) (json.RawMessage, time.Time, error) {
	var answers json.RawMessage
	var createdAt time.Time

	err := db.QueryRow("SELECT answers, created_at FROM user_session WHERE user_id = $1 AND signature = $2",
		userID, signature).Scan(&answers, &createdAt)
	if err != nil {
		return nil, time.Time{}, err
	}

	return answers, createdAt, nil
}

func SaveSession(userId string, signature string, answers []string) error {
	jsonAnswers, err := json.Marshal(answers)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO user_session (user_id, signature, answers) VALUES ($1, $2, $3)",
		userId, signature, jsonAnswers)
	return err
}
