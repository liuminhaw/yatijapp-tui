package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ResetPasswordTokenRequest struct {
	Email string `json:"email"`
}

func (r ResetPasswordTokenRequest) Do(serverUrl string) (Message, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return Message{}, LoadApiDataErr{
			Err: err,
			Msg: "Failed to create request body JSON",
		}
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/v1/tokens/password-reset", serverUrl),
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return Message{}, LoadApiDataErr{
			Err: err,
			Msg: "API request error: POST ResetPasswordToken",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return Message{}, LoadApiDataErr{
				Status: resp.StatusCode,
				Err:    err,
				Msg:    "API error response decode failure",
			}
		}

		return Message{}, LoadApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	var responseData Message
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return Message{}, LoadApiDataErr{
			Status: resp.StatusCode,
			Err:    err,
			Msg:    "API response decode error",
		}
	}

	return responseData, nil
}
