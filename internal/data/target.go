package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
)

type ListTargetsResponse struct {
	Metadata Metadata `json:"metadata"`
	Targets  []Target `json:"targets"`
	Error    string   `json:"error,omitempty"`
}

func ListTargets(serverURL string, client *authclient.AuthClient) ([]Target, error) {
	// time.Sleep(3 * time.Second) // Simulate a delay for loading targets

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/v1/targets", serverURL),
		nil,
	)
	if err != nil {
		return []Target{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create GET request for Targets",
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return []Target{}, respErrorCheck(err, "API request error: GET Targets")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return []Target{}, UnexpectedApiDataErr{
				Err: err,
				Msg: "API error response decode failure",
			}
		}

		return []Target{}, UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	var responseData ListTargetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return []Target{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "API response decode error",
		}
	}

	return responseData.Targets, nil
}

type GetTargetResponse struct {
	Target Target `json:"target"`
	Error  string `json:"error,omitempty"`
}

func GetTarget(serverURL, uuid string, client *authclient.AuthClient) (Target, error) {
	// time.Sleep(2 * time.Second) // Simulate a delay for loading targets

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/v1/targets/%s", serverURL, uuid),
		nil,
	)
	if err != nil {
		return Target{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create GET request for Target",
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return Target{}, respErrorCheck(err, "API request error: GET Target")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return Target{}, UnexpectedApiDataErr{
				Err: err,
				Msg: "API error response decode failure",
			}
		}

		return Target{}, UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	var responseData GetTargetResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return Target{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "API response decode error",
		}
	}

	responseData.Target.CreatedAt = responseData.Target.CreatedAt.Local()
	responseData.Target.UpdatedAt = responseData.Target.UpdatedAt.Local()
	responseData.Target.LastActive = responseData.Target.LastActive.Local()

	return responseData.Target, nil
}

func DeleteTarget(serverURL, uuid string, client *authclient.AuthClient) error {
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/v1/targets/%s", serverURL, uuid),
		nil,
	)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create DELETE request for target",
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return respErrorCheck(err, "API request error: DELETE Target")
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

		return UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	return nil
}

type TargetRequestBody struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     *Date  `json:"due_date,omitempty"`
	Notes       string `json:"notes"`
	Status      string `json:"status"`
}

func (b TargetRequestBody) Create(serverURL string, client *authclient.AuthClient) error {
	data, err := json.Marshal(b)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create request body JSON",
		}
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/v1/targets", serverURL),
		bytes.NewBuffer(data),
	)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create POST request for target",
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return respErrorCheck(err, "API request error: POST Target")
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
			Msg:    responseErr.Error(),
		}
	}

	return nil
}

func (b TargetRequestBody) Update(
	serverURL string,
	uuid string,
	client *authclient.AuthClient,
) error {
	data, err := json.Marshal(b)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create request body JSON",
		}
	}

	req, err := http.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("%s/v1/targets/%s", serverURL, uuid),
		bytes.NewBuffer(data),
	)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create PATCH request for target",
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return respErrorCheck(err, "API request error: PATCH Target")
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

		return UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	return nil
}
