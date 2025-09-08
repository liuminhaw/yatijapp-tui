package data

import (
	"bytes"
	"encoding/json"
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
		return User{}, UnauthorizedApiDataErr{
			Err: err,
			Msg: "Failed to create GET request for Current User",
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return User{}, respErrorCheck(err, "API request error: GET Current User")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return User{}, UnauthorizedApiDataErr{
				Status: resp.StatusCode,
				Err:    err,
				Msg:    "API error response decode failure",
			}
		}

		return User{}, UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	var responseData GetUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return User{}, UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    err,
			Msg:    "API response decode error",
		}
	}

	return responseData.User, nil
}

type UserRequest struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r UserRequest) Signin(serverURL string, tokenPath string) error {
	data, err := json.Marshal(r)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create request body JSON",
		}
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/v1/tokens/authentication", serverURL),
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "API request error: POST Signin",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return UnexpectedApiDataErr{
				Err: err,
				Msg: "API error response decode failure",
			}
		}

		return UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    "Incorrect email or password",
		}
	}

	var responseData struct {
		AuthToken authclient.Token `json:"authentication_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "API response decode error",
		}
	}
	if err := authclient.TokenWrite(responseData.AuthToken, tokenPath); err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to generate user token",
		}
	}

	return nil
}

func (r UserRequest) Register(serverURL string) error {
	data, err := json.Marshal(r)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create request body JSON",
		}
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/v1/users", serverURL),
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "API request error: POST Register",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return UnexpectedApiDataErr{
				Err: err,
				Msg: "API error response decode failure",
			}
		}

		return UnmatchedApiRespDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	return nil
}

type UserTokenRequest struct {
	Token    string `json:"token"`
	Password string `json:"password,omitempty"`
}

func (r UserTokenRequest) ResetPassword(serverURL string) (Message, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return Message{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create request body JSON",
		}
	}

	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/v1/users/password", serverURL),
		bytes.NewReader(data),
	)
	if err != nil {
		return Message{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create PUT request for Reset Password",
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Message{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "API request error: PUT ResetPassword",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return Message{}, UnexpectedApiDataErr{
				Err: err,
				Msg: "API error response decode failure",
			}
		}

		return Message{}, UnmatchedApiRespDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	var responseData Message
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return Message{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "API response decode error",
		}
	}

	return responseData, nil
}

func (r UserTokenRequest) ActivateUser(serverURL string) error {
	data, err := json.Marshal(r)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create request body JSON",
		}
	}

	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/v1/users/activated", serverURL),
		bytes.NewReader(data),
	)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create PUT request for Activating User",
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "API request error: PUT ActivateUser",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return UnexpectedApiDataErr{
				Err: err,
				Msg: "API error response decode failure",
			}
		}

		return UnmatchedApiRespDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	return nil
}
