package main

import (
	"log"
	"net/http"
	"test-signer/internal/signer"
)

func main() {
	http.HandleFunc("/sign", signer.SignAnswers)
	http.HandleFunc("/verify", signer.VerifySignature)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
