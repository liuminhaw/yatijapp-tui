package data

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
)

type ListActionsResponse struct {
	Metadata Metadata `json:"metadata"`
	Actions  []Action `json:"actions"`
	Error    string   `json:"error,omitempty"`
}

func ListActions(
	info ListRequestInfo,
	client *authclient.AuthClient,
) (ListActionsResponse, error) {
	var reqUrl string
	var err error
	if info.SrcUUID == "" {
		reqUrl, err = info.requestUrl("/v1/actions")
	} else {
		reqUrl, err = info.requestUrl("/v1/targets/" + info.SrcUUID + "/actions")
	}
	if err != nil {
		return ListActionsResponse{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create request URL for GET Actions",
		}
	}

	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return ListActionsResponse{}, UnauthorizedApiDataErr{
			Err: err,
			Msg: "Failed to create GET request for Actions",
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return ListActionsResponse{}, respErrorCheck(err, "API request error: GET Actions")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return ListActionsResponse{}, UnexpectedApiDataErr{
				Err: err,
				Msg: "API error response decode failure",
			}
		}

		return ListActionsResponse{}, UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	var responseData ListActionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return ListActionsResponse{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "API response decode error",
		}
	}

	return responseData, nil
}

type GetActionResponse struct {
	Action Action `json:"action"`
	Error  string `json:"error,omitempty"`
}

func GetAction(serverURL, uuid string, client *authclient.AuthClient) (Action, error) {
	path, err := url.JoinPath(serverURL, "v1", "actions", uuid)
	if err != nil {
		return Action{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create URL path for GET Action",
		}
	}

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return Action{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create GET request for Action",
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return Action{}, respErrorCheck(err, "API request error: GET Action")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return Action{}, UnexpectedApiDataErr{
				Err: err,
				Msg: "API error response decode failure",
			}
		}

		return Action{}, UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	var responseData GetActionResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return Action{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "API response decode error",
		}
	}

	responseData.Action.CreatedAt = responseData.Action.CreatedAt.Local()
	responseData.Action.UpdatedAt = responseData.Action.UpdatedAt.Local()
	responseData.Action.LastActive = responseData.Action.LastActive.Local()

	return responseData.Action, nil
}

func DeleteAction(serverURL, uuid string, client *authclient.AuthClient) error {
	path, err := url.JoinPath(serverURL, "v1", "actions", uuid)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create URL path for DELETE Action",
		}
	}

	req, err := http.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create DELETE request for Action",
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return respErrorCheck(err, "API request error: DELETE Action")
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

type ActionRequestBody struct {
	TargetUUID  string `json:"target_uuid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     *Date  `json:"due_date,omitempty"`
	Notes       string `json:"notes"`
	Status      string `json:"status"`
}

func (b ActionRequestBody) Create(serverURL string, client *authclient.AuthClient) error {
	path, err := url.JoinPath(serverURL, "v1", "actions")
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create URL path for POST Action",
		}
	}
	data, err := json.Marshal(b)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create request body JSON",
		}
	}

	req, err := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(data))
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create POST request for Action",
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return respErrorCheck(err, "API request error: POST Action")
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

func (b ActionRequestBody) Update(
	serverURL, uuid string, client *authclient.AuthClient,
) error {
	path, err := url.JoinPath(serverURL, "v1", "actions", uuid)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create URL path for PATCH Action",
		}
	}

	data, err := json.Marshal(b)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create request body JSON",
		}
	}

	req, err := http.NewRequest(http.MethodPatch, path, bytes.NewBuffer(data))
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create PATCH request for Action",
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return respErrorCheck(err, "API request error: PATCH Action")
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
