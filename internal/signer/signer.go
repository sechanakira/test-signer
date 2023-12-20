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
	var req SignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	signature := generateSignature(req)

	saveSession(req.UserID, signature, req.Answers)

	resp := SignResponse{Signature: signature}

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
	db, _ := conn.InitDB()
	jsonAnswers, err := json.Marshal(answers)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO user_session (user_id, signature, answers) VALUES ($1, $2, $3)",
		userId, signature, jsonAnswers)
	return err
}
