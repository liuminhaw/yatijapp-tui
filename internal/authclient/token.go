package authclient

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	SessionUUID  string `json:"session_id"`
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
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

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

func TokenDelete(path string) error {
	err := os.Remove(path)
	if err == nil || os.IsNotExist(err) {
		return nil
	}
	return err
}
