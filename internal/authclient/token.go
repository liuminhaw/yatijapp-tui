package authclient

import (
	"encoding/json"
	"os"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func TokenRead(path string) (Token, error) {
	file, err := os.Open(path)
	if err != nil {
		return Token{}, err
	}

	var token Token
	if err := json.NewDecoder(file).Decode(&token); err != nil {
		return Token{}, err
	}

	return token, nil
}

func TokenWrite(t Token, path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(t); err != nil {
		return err
	}

	return nil
}
