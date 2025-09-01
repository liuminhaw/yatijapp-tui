package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
)

func Signout(serverURL string, client *authclient.AuthClient) (Message, error) {
	token, err := client.GetToken()
	if err != nil {
		return Message{}, LoadApiDataErr{
			Err: err,
			Msg: "Failed to read user token",
		}
	}

	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/v1/sessions/%s", serverURL, token.SessionUUID),
		nil,
	)
	if err != nil {
		return Message{}, LoadApiDataErr{
			Err: err,
			Msg: "Failed to create DELETE request for Sign out",
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		if errors.As(err, &authclient.ErrInvalidToken{}) {
			return Message{}, LoadApiDataErr{
				Status: http.StatusUnauthorized,
				Err:    err,
				Msg:    err.Error(),
			}
		}
		return Message{}, LoadApiDataErr{
			Err: err,
			Msg: "API request error: DELETE Sign out",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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
	// It's ok if clear token fails, because token is invalid after sign out
	client.ClearToken()

	return responseData, nil
}
