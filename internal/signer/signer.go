package signer

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	conn "test-signer/internal/db"
	"time"
)

func SignAnswers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed here", http.StatusMethodNotAllowed)
		return
	}

	var req SignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		http.Error(w, "user id cannot be null", http.StatusBadRequest)
		return
	}

	if req.Answers == nil {
		http.Error(w, "answers are required", http.StatusBadRequest)
		return
	}

	signature := generateSignature(req)

	err := saveSession(req.UserID, signature, req.Answers)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := SignResponse{Signature: signature}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(resp)
}

func VerifySignature(w http.ResponseWriter, r *http.Request) {
	db, _ := conn.InitDB()

	var req struct {
		UserID    string `json:"userId"`
		Signature string `json:"signature"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var answers json.RawMessage
	var createdAt time.Time
	err := db.QueryRow("SELECT answers, created_at FROM user_session WHERE user_id = $1 AND signature = $2", req.UserID, req.Signature).Scan(&answers, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No matching session found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
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
	db, dbErr := conn.InitDB()
	if dbErr != nil {
		return dbErr
	}

	jsonAnswers, err := json.Marshal(answers)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO user_session (user_id, signature, answers) VALUES ($1, $2, $3)",
		userId, signature, jsonAnswers)
	return err
}
