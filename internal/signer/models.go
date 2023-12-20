package signer

type SignRequest struct {
	UserID  string   `json:"userId"`
	Answers []string `json:"answers"`
}

type SignResponse struct {
	Signature string `json:"signature"`
}
