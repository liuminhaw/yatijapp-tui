package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
)

type GetUserResponse struct {
	User User `json:"user"`
}

func GetCurrentUser(serverURL string, client *authclient.AuthClient) (User, error) {
	// time.Sleep(3 * time.Second) // Simulate a delay for loading targets

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/v1/users/me", serverURL),
		nil,
	)
	if err != nil {
		return User{}, LoadApiDataErr{
			Err: err,
			Msg: "Failed to create GET request for Current User",
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		if errors.As(err, &authclient.ErrInvalidToken{}) {
			return User{}, LoadApiDataErr{
				Status: http.StatusUnauthorized,
				Err:    err,
				Msg:    err.Error(),
			}
		}
		return User{}, LoadApiDataErr{
			Err: err,
			Msg: "API request error: GET Current User",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return User{}, LoadApiDataErr{
				Status: resp.StatusCode,
				Err:    err,
				Msg:    "API error response decode failure",
			}
		}

		return User{}, LoadApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	var responseData GetUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return User{}, LoadApiDataErr{
			Status: resp.StatusCode,
			Err:    err,
			Msg:    "API response decode error",
		}
	}

	return responseData.User, nil
}
