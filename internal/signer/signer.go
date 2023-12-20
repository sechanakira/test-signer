package signer

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	conn "test-signer/internal/db"
	"time"
)

var db *sql.DB

func init() {
	var err error
	db, err = conn.InitDB()
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

func SignAnswers(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("failed to close response body")
		}
	}(r.Body)

	if err := validateJwtPresent(r.Header.Get("Authorization")); err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if err := validateHttpMethod(r.Method, "POST"); err != nil {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed here")
		return
	}

	var req SignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := validateRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	signature := generateSignature(req)
	if err := saveSession(req.UserID, signature, req.Answers); err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSONResponse(w, http.StatusCreated, SignResponse{Signature: signature})
}

func validateHttpMethod(method string, allowedMethod string) error {
	if method != allowedMethod {
		return fmt.Errorf("method not allowed")
	}
	return nil
}

func validateRequest(req SignRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user id cannot be null")
	}
	if len(req.Answers) == 0 {
		return fmt.Errorf("answers are required")
	}
	return nil
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(map[string]string{"error": message})
	if err != nil {
		log.Printf("failed to write error %s ", err)
		return
	}
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("failed to write json response %s", err)
		return
	}
}

func VerifySignature(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close response body %s ", err)
		}
	}(r.Body)

	if err := validateHttpMethod(r.Method, "POST"); err != nil {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed here")
		return
	}

	var req struct {
		UserID    string `json:"userId"`
		Signature string `json:"signature"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := validateVerifyRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	answers, createdAt, err := fetchSessionData(req.UserID, req.Signature)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "No matching session found")
		} else {
			writeError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	response := struct {
		Status    string          `json:"status"`
		Answers   json.RawMessage `json:"answers"`
		Timestamp time.Time       `json:"timestamp"`
	}{
		Status:    "OK",
		Answers:   answers,
		Timestamp: createdAt,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func validateVerifyRequest(req struct {
	UserID    string `json:"userId"`
	Signature string `json:"signature"`
}) error {
	if req.UserID == "" {
		return fmt.Errorf("userId required")
	}
	if req.Signature == "" {
		return fmt.Errorf("signature required")
	}
	return nil
}

func fetchSessionData(userID, signature string) (json.RawMessage, time.Time, error) {
	var answers json.RawMessage
	var createdAt time.Time

	err := db.QueryRow("SELECT answers, created_at FROM user_session WHERE user_id = $1 AND signature = $2",
		userID, signature).Scan(&answers, &createdAt)
	if err != nil {
		return nil, time.Time{}, err
	}

	return answers, createdAt, nil
}

func generateSignature(req SignRequest) string {
	hash := sha256.New()
	hash.Write([]byte(req.UserID))
	for _, ans := range req.Answers {
		hash.Write([]byte(ans))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func saveSession(userId string, signature string, answers []string) error {
	jsonAnswers, err := json.Marshal(answers)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO user_session (user_id, signature, answers) VALUES ($1, $2, $3)",
		userId, signature, jsonAnswers)
	return err
}

func validateJwtPresent(authHeader string) error {
	if authHeader == "" {
		return fmt.Errorf("authorization header is required")
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return fmt.Errorf("authorization header format must be Bearer {token}")
	}

	return nil
}
