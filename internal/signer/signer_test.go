package signer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateHttpMethod(t *testing.T) {
	err := validateHttpMethod("POST", "POST")
	assert.NoError(t, err)

	err = validateHttpMethod("GET", "POST")
	assert.Error(t, err)
}

func TestValidateVerifyRequest(t *testing.T) {
	req := struct {
		UserID    string `json:"userId"`
		Signature string `json:"signature"`
	}{"user1", "signature1"}
	assert.NoError(t, validateVerifyRequest(req))

	req = struct {
		UserID    string `json:"userId"`
		Signature string `json:"signature"`
	}{"", "signature1"}
	assert.Error(t, validateVerifyRequest(req))

	req = struct {
		UserID    string `json:"userId"`
		Signature string `json:"signature"`
	}{"user1", ""}
	assert.Error(t, validateVerifyRequest(req))
}

func TestValidateJwtPresent(t *testing.T) {
	assert.Error(t, validateJwtPresent(""))
	assert.Error(t, validateJwtPresent("Basic token"))
	assert.NoError(t, validateJwtPresent("Bearer token"))
}

func TestValidateRequest(t *testing.T) {
	assert.Error(t, validateRequest(SignRequest{UserID: "", Answers: []string{"answer1"}}))
	assert.Error(t, validateRequest(SignRequest{UserID: "user1", Answers: []string{}}))
	assert.NoError(t, validateRequest(SignRequest{UserID: "user1", Answers: []string{"answer1"}}))
}
